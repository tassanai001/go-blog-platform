# Go Blog Platform

A RESTful blog platform built with Go, featuring role-based access control and MongoDB integration.

## Features

- User authentication with JWT
- Role-based access control (Admin, Author, Reader)
- Blog post management
- MongoDB integration
- RESTful API design
- Media management

## Prerequisites

- Go 1.21 or higher
- MongoDB 4.4 or higher
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/tassanai001/go-blog-platform.git
cd go-blog-platform
```

2. Install dependencies:
```bash
go mod download
```

3. Configure MongoDB:
- Make sure MongoDB is running on localhost:27017
- Update the configuration in `config/config.go` if needed

4. Run the application:
```bash
go run cmd/server/main.go
```

## API Endpoints

### Authentication
- `POST /api/register` - Register a new user
- `POST /api/login` - Login and get JWT token

### Posts (Protected Routes)
- `GET /api/posts` - List all posts
- `GET /api/posts/:id` - Get a specific post
- `POST /api/posts` - Create a new post (Author, Admin)
- `PUT /api/posts/:id` - Update a post (Author, Admin)
- `DELETE /api/posts/:id` - Delete a post (Admin)

### Post Management with Media

Posts can include a featured image and a gallery of images. When creating or updating a post, you can include media files:

```bash
# Create a post with media
POST /api/posts
Content-Type: multipart/form-data
Authorization: Bearer YOUR_JWT_TOKEN

Form Data:
- title: Post title
- content: Post content
- status: published/draft
- tags[]: tag1, tag2, etc.
- featured_image: Single image file
- gallery[]: Multiple image files
```

### Author Routes
- `GET /api/author/drafts` - List author's drafts
- `POST /api/author/drafts` - Create a draft

### Admin Routes
- `GET /api/admin/users` - List all users
- `PUT /api/admin/users/:id/role` - Update user role
- `DELETE /api/admin/users/:id` - Delete user

### User Profile with Media

User profiles support avatar and cover images. These can be set during registration or updated later:

#### Register with Profile Images
```bash
POST /api/auth/register
Content-Type: multipart/form-data

Form Data:
- username: Username
- password: Password
- email: user@example.com
- full_name: Full Name
- avatar: Profile picture
- cover_image: Profile cover image
- bio: User bio
- location: User location
- website: Personal website
- social_links[twitter]: Twitter handle
- social_links[github]: GitHub username
```

#### Update Profile with Images
```bash
PUT /api/users/profile
Content-Type: multipart/form-data
Authorization: Bearer YOUR_JWT_TOKEN

Form Data:
- full_name: Updated full name
- avatar: New profile picture
- cover_image: New cover image
- bio: Updated bio
- location: Updated location
- website: Updated website
- social_links[twitter]: Updated Twitter handle
- social_links[github]: Updated GitHub username
```

### Image Specifications

- Supported formats: JPEG, PNG, GIF, WebP
- Maximum file size: 10MB
- Thumbnails are automatically generated in three sizes:
  - Small: 150x150
  - Medium: 300x300
  - Large: 600x600

### Security Considerations

- All uploads require authentication
- File types are validated using MIME detection
- Files are stored in user-specific directories
- Original filenames are sanitized
- Secure file paths are enforced
- Old media files are automatically deleted when replaced

### Media Management

The blog platform includes a comprehensive media management system that supports image uploads, automatic thumbnail generation, and metadata management.

#### Features

- Image upload and storage
- Automatic thumbnail generation (small: 150x150, medium: 300x300, large: 600x600)
- File type validation (JPEG, PNG, GIF, WebP)
- Image optimization
- Secure file storage
- Media metadata management

#### API Endpoints

##### Upload Media
```bash
POST /api/media/upload
Content-Type: multipart/form-data
Authorization: Bearer YOUR_JWT_TOKEN

Form Data:
- file: File to upload (max 10MB)
```

##### List Media
```bash
GET /api/media/list
Authorization: Bearer YOUR_JWT_TOKEN
```

##### Get Media Details
```bash
GET /api/media/:id
Authorization: Bearer YOUR_JWT_TOKEN
```

##### Update Media Metadata
```bash
PUT /api/media/:id
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN

{
  "title": "Image Title",
  "description": "Image description",
  "alt_text": "Alternative text",
  "tags": ["nature", "landscape"]
}
```

##### Delete Media
```bash
DELETE /api/media/:id
Authorization: Bearer YOUR_JWT_TOKEN
```

#### Example Usage

1. Upload an image:
```bash
curl -X POST http://localhost:8080/api/media/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/image.jpg"
```

2. List all media:
```bash
curl -X GET http://localhost:8080/api/media/list \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

3. Update metadata:
```bash
curl -X PUT http://localhost:8080/api/media/MEDIA_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Sunset",
    "description": "Beautiful sunset at the beach",
    "alt_text": "Orange sun setting over ocean waves",
    "tags": ["sunset", "beach", "nature"]
  }'
```

#### Response Format

```json
{
  "id": "media_id",
  "user_id": "user_id",
  "file_name": "image.jpg",
  "file_type": ".jpg",
  "mime_type": "image/jpeg",
  "size": 1024,
  "path": "path/to/file.jpg",
  "url": "http://localhost:8080/media/path/to/file.jpg",
  "thumbnails": [
    {
      "size": "small",
      "width": 150,
      "height": 150,
      "path": "path/to/small.jpg",
      "url": "http://localhost:8080/media/path/to/small.jpg"
    },
    // medium and large thumbnails...
  ],
  "metadata": {
    "width": 1920,
    "height": 1080,
    "title": "Image Title",
    "description": "Image description",
    "alt_text": "Alternative text",
    "tags": ["tag1", "tag2"]
  },
  "created_at": "2025-02-22T15:04:05Z",
  "updated_at": "2025-02-22T15:04:05Z"
}
```

#### Configuration

The media system is configured through environment variables:

```env
# Base URL for media access
BASE_URL=http://localhost:8080

# Maximum file size (in bytes, default: 10MB)
MAX_FILE_SIZE=10485760

# Allowed file types
ALLOWED_FILE_TYPES=image/jpeg,image/png,image/gif,image/webp
```

#### Security Considerations

- Files are stored in user-specific directories
- File types are validated using MIME type detection
- File size is limited to prevent abuse
- Thumbnails are generated asynchronously
- Original filenames are sanitized
- Secure file paths are enforced

## User Roles

1. Reader (Default)
   - Can view posts
   - Basic access to public content

2. Author
   - All Reader permissions
   - Can create and edit posts
   - Can manage drafts

3. Admin
   - All Author permissions
   - Full system access
   - User management
   - Content moderation

## License

MIT License
