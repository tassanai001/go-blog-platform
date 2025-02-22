# Go Blog Platform

A RESTful blog platform built with Go, featuring role-based access control and MongoDB integration.

## Features

- User authentication with JWT
- Role-based access control (Admin, Author, Reader)
- Blog post management
- MongoDB integration
- RESTful API design

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

### Author Routes
- `GET /api/author/drafts` - List author's drafts
- `POST /api/author/drafts` - Create a draft

### Admin Routes
- `GET /api/admin/users` - List all users
- `PUT /api/admin/users/:id/role` - Update user role
- `DELETE /api/admin/users/:id` - Delete user

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
