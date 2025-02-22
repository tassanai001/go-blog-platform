package services

import (
	"fmt"
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
	uploadDir string
	baseURL   string
}

type ThumbnailInfo struct {
	Size string
	Path string
}

func NewMediaService(uploadDir string, baseURL string) *MediaService {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}

	return &MediaService{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

func (s *MediaService) ValidateFile(file *multipart.FileHeader) error {
	// Check file size (10MB limit)
	if file.Size > 10*1024*1024 {
		return fmt.Errorf("file size exceeds 10MB limit")
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read the first 512 bytes to detect content type
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
	mime := mimetype.Detect(buffer)
	if !strings.HasPrefix(mime.String(), "image/") {
		return fmt.Errorf("invalid file type: only images are allowed")
	}

	return nil
}

func (s *MediaService) SaveFile(file *multipart.FileHeader, userID string) (string, []ThumbnailInfo, error) {
	// Create user directory if it doesn't exist
	userDir := filepath.Join(s.uploadDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", nil, err
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(userDir, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", nil, err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", nil, err
	}
	defer dst.Close()

	// Copy file contents
	if _, err = io.Copy(dst, src); err != nil {
		return "", nil, err
	}

	// Generate thumbnails for images
	thumbnails, err := s.generateThumbnails(filePath, userID)
	if err != nil {
		// Log the error but don't fail the upload
		fmt.Printf("Failed to generate thumbnails: %v\n", err)
	}

	return filePath, thumbnails, nil
}

func (s *MediaService) generateThumbnails(sourcePath string, userID string) ([]ThumbnailInfo, error) {
	// Open the source image
	src, err := imaging.Open(sourcePath)
	if err != nil {
		return nil, err
	}

	// Define thumbnail sizes
	sizes := map[string]int{
		"small":  150,
		"medium": 300,
		"large":  600,
	}

	var thumbnails []ThumbnailInfo

	// Generate thumbnails for each size
	for size, dimension := range sizes {
		// Create thumbnail filename
		ext := filepath.Ext(sourcePath)
		filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), size, ext)
		thumbPath := filepath.Join(s.uploadDir, userID, "thumbnails", filename)

		// Create thumbnails directory if it doesn't exist
		thumbDir := filepath.Dir(thumbPath)
		if err := os.MkdirAll(thumbDir, 0755); err != nil {
			return thumbnails, err
		}

		// Resize image
		thumb := imaging.Resize(src, dimension, dimension, imaging.Lanczos)

		// Save thumbnail
		err = imaging.Save(thumb, thumbPath)
		if err != nil {
			return thumbnails, err
		}

		thumbnails = append(thumbnails, ThumbnailInfo{
			Size: size,
			Path: thumbPath,
		})
	}

	return thumbnails, nil
}

func (s *MediaService) DeleteFile(path string) error {
	// Delete the main file
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Try to delete thumbnails
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Delete thumbnail directory
	thumbDir := filepath.Join(dir, "thumbnails")
	entries, err := os.ReadDir(thumbDir)
	if err != nil {
		// Ignore if thumbnail directory doesn't exist
		return nil
	}

	// Delete all thumbnails that start with the same name
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), name+"_") {
			thumbPath := filepath.Join(thumbDir, entry.Name())
			_ = os.Remove(thumbPath)
		}
	}

	return nil
}
