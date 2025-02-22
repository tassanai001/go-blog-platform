package services

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
)

type MediaService struct {
	uploadDir      string
	allowedTypes   []string
	maxFileSize    int64
	thumbnailSizes map[string]struct{ width, height int }
}

func NewMediaService(uploadDir string) *MediaService {
	return &MediaService{
		uploadDir: uploadDir,
		allowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		maxFileSize: 10 * 1024 * 1024, // 10MB
		thumbnailSizes: map[string]struct{ width, height int }{
			"small":  {width: 150, height: 150},
			"medium": {width: 300, height: 300},
			"large":  {width: 600, height: 600},
		},
	}
}

// ValidateFile checks if the file is valid
func (s *MediaService) ValidateFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > s.maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.maxFileSize)
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read the first 512 bytes to detect the content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	// Reset the read pointer
	_, err = src.Seek(0, 0)
	if err != nil {
		return err
	}

	// Detect MIME type
	mtype := mimetype.Detect(buffer)
	allowed := false
	for _, t := range s.allowedTypes {
		if mtype.String() == t {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("file type %s is not allowed", mtype.String())
	}

	return nil
}

// SaveFile saves the uploaded file and creates thumbnails if it's an image
func (s *MediaService) SaveFile(file *multipart.FileHeader, userID string) (string, []struct{ Size, Path string }, error) {
	// Create user upload directory if it doesn't exist
	userDir := filepath.Join(s.uploadDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", nil, err
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filepath := filepath.Join(userDir, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", nil, err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", nil, err
	}
	defer dst.Close()

	// Copy the uploaded file
	if _, err = io.Copy(dst, src); err != nil {
		return "", nil, err
	}

	// If it's an image, create thumbnails
	thumbnails := []struct{ Size, Path string }{}
	if isImage := s.isImageFile(file.Filename); isImage {
		thumbnails, err = s.createThumbnails(filepath, userDir, filename)
		if err != nil {
			// Log the error but don't fail the upload
			fmt.Printf("Error creating thumbnails: %v\n", err)
		}
	}

	return filepath, thumbnails, nil
}

// isImageFile checks if the file is an image based on extension
func (s *MediaService) isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
}

// createThumbnails generates thumbnails for the image
func (s *MediaService) createThumbnails(originalPath, userDir, filename string) ([]struct{ Size, Path string }, error) {
	// Open the original image
	src, err := imaging.Open(originalPath)
	if err != nil {
		return nil, err
	}

	thumbnails := []struct{ Size, Path string }{}
	baseFilename := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Create thumbnails for each size
	for size, dimensions := range s.thumbnailSizes {
		thumbnailFilename := fmt.Sprintf("%s_%s%s", baseFilename, size, filepath.Ext(filename))
		thumbnailPath := filepath.Join(userDir, thumbnailFilename)

		// Resize the image
		thumbnail := imaging.Fit(src, dimensions.width, dimensions.height, imaging.Lanczos)

		// Save the thumbnail
		err = imaging.Save(thumbnail, thumbnailPath)
		if err != nil {
			return thumbnails, err
		}

		thumbnails = append(thumbnails, struct{ Size, Path string }{Size: size, Path: thumbnailPath})
	}

	return thumbnails, nil
}

// DeleteFile removes a file and its thumbnails
func (s *MediaService) DeleteFile(filePath string) error {
	// Delete the original file
	if err := os.Remove(filePath); err != nil {
		return err
	}

	// If it's an image, delete thumbnails
	if s.isImageFile(filePath) {
		dirPath := filepath.Dir(filePath)
		fileName := filepath.Base(filePath)
		fileExt := filepath.Ext(fileName)
		baseName := strings.TrimSuffix(fileName, fileExt)

		for size := range s.thumbnailSizes {
			thumbnailPath := filepath.Join(dirPath, fmt.Sprintf("%s_%s%s", baseName, size, fileExt))
			if err := os.Remove(thumbnailPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete thumbnail %s: %w", thumbnailPath, err)
			}
		}
	}

	return nil
}

// OptimizeImage compresses and optimizes the image
func (s *MediaService) OptimizeImage(path string) error {
	// Open the image
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Create a new file for the optimized image
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Encode with optimization settings based on format
	switch format {
	case "jpeg":
		return jpeg.Encode(out, img, &jpeg.Options{Quality: 85})
	case "png":
		return png.Encode(out, img)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
}
