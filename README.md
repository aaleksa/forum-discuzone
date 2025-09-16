# 🌐 Go Web Forum

A modern web forum built with Go, SQLite, and Docker that allows users to communicate through posts and comments with category-based organization and like/dislike functionality.

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Project Structure](#-project-structure)
- [Database Schema](#-database-schema)
- [Design Specification](#-design-specification)
- [Installation](#-installation)
- [Usage](#-usage)
- [Development Plan](#-development-plan)
- [API Endpoints](#-api-endpoints)
- [Contributing](#-contributing)

## ✨ Features

- **User Management**
  - User registration and authentication
  - Secure password hashing with bcrypt
  - Session management with UUID tokens
  - User profiles with activity tracking

- **Posts & Comments**
  - Create, read posts with rich content
  - Comment system with threaded discussions
  - Category-based post organization
  - Post filtering and search functionality

- **Interaction System**
  - Like/Dislike posts and comments
  - Real-time reaction counting
  - User engagement tracking

- **Filtering & Search**
  - Filter by categories
  - Filter by user's created posts
  - Filter by user's liked posts
  - Sort by popularity and date

## 🛠 Tech Stack

- **Backend**: Go (Golang)
- **Database**: SQLite
- **Authentication**: Sessions with UUID, bcrypt password hashing
- **Containerization**: Docker
- **Frontend**: HTML templates, CSS, JavaScript

## 📂 Project Structure

```
/forum
├── database/
│   ├── db_setup.go
│   └── forum.db
├── internal/
│   ├── handlers/
│   │   ├── comments.go
│   │   ├── create-post.go
│   │   ├── create-reactions.go
│   │   ├── get-categories.go
│   │   ├── get-likes-count.go
│   │   ├── get-post.go
│   │   ├── get-posts-list.go
│   │   ├── get-reactions.go
│   │   ├── login.go
│   │   ├── logout.go
│   │   ├── profile.go
│   │   ├── registration.go
│   │   └── user.go
│   └── utils/
│       ├── category.go
│       └── open_db.go
├── static/
│   ├── css/
│   └── img/
├── templates/
│   ├── 400.html
│   ├── 401.html
│   ├── 404.html
│   ├── 405.html
│   ├── 409.html
│   ├── 500.html
│   ├── add_comment.html
│   ├── comment.html
│   ├── create_post.html
│   ├── filters.html
│   ├── header.html
│   ├── index.html
│   ├── layout.html
│   ├── login.html
│   ├── nav.html
│   ├── post_item.html
│   ├── post_list.html
│   ├── post_list_item.html
│   ├── post_page.html
│   ├── profile.html
│   ├── registration.html
│   ├── success.html
│   ├── user_page.html
│   ├── view_posts.html
│   ├── categories.txt
│   ├── design-ua.txt
│   └── design-uk.txt
├── error.go
├── main.go
├── go.mod
├── go.sum
├── Dockerfile
├── project_description.md
├── README.md
└── test.txt
```

## 🗄 Database Schema

### Users Table
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    role TEXT DEFAULT 'user',
    avatar_url TEXT DEFAULT NULL,
    banned BOOLEAN DEFAULT FALSE,
    provider TEXT DEFAULT '',
    provider_id TEXT DEFAULT ''
);
```

### Posts Table
```sql
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    image_path TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

```

### Comments Table
```sql
CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (post_id) REFERENCES posts(id)
);
```

### Categories Table
```sql
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);
```

### Post Categories Junction Table
```sql
CREATE TABLE post_categories (
    post_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

```

### Likes/Dislikes Table
```sql
CREATE TABLE likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    post_id INTEGER,
    comment_id INTEGER,
    reaction TEXT NOT NULL CHECK (reaction IN ('Like', 'Dislike')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (post_id IS NOT NULL OR comment_id IS NOT NULL),
    UNIQUE(user_id, post_id, comment_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
);
```
### Post Images
```sql
CREATE TABLE post_images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    image_path TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    order_index INTEGER DEFAULT 0,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);
```
### Tags & Post-Tags
```sql
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE post_tags (
    post_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, tag_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX idx_tags_name ON tags(name);

```
### Sessions
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```
### Moderator Requests
```sql
CREATE TABLE moderator_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    status TEXT CHECK(status IN ('pending', 'approved', 'rejected')) DEFAULT 'pending',
    requested_at TEXT NOT NULL,
    reviewed_at TEXT,
    reviewed_by INTEGER,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (reviewed_by) REFERENCES users(id)
);
```
### Notifications
```sql
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('like', 'dislike', 'comment', 'reply')),
    post_id INTEGER,
    comment_id INTEGER,
    actor_id INTEGER,
    is_read BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (comment_id) REFERENCES comments(id),
    FOREIGN KEY (actor_id) REFERENCES users(id)
);
```
### Password Resets
```sql
CREATE TABLE password_resets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```
## Open DB in terminal
 ``` sqlcipher forum.db ```
 ``` PRAGMA key = 'discuzoneForumZone1281'; ```

## 🎨 Design Specification

### Core Design Principles
- **Minimalism**: Clean, simple interface with no unnecessary elements
- **User-Friendly Navigation**: Easy access to main features
- **Responsive Design**: Optimized for desktop, tablet, and smartphone
- **Intuitive UX**: Clear interaction patterns

### Color Scheme (Light Mode)
- **Background**: `#F8F9FA` (light gray/near white)
- **Text**: `#333333` (dark gray)
- **Accent**: `#4C9A2A`, `#76BA1B` (green shades)
- **Links**: `#007bff` (blue)
- **Buttons**: `#4C9A2A` (hover: `#76BA1B`)
- **Tags**: `#A4DE02`
- **Meta text**: `#888888` (gray)

### Page Layout Examples

#### Home Page
```
# Forum Name [Logo]

[Search] [Register] [Login] [Create Post]

## Filters
- [Categories] [Popular] [Newest]

## Posts List
### Post Title 1
**Author:** User1 | **Date:** 01.01.2023
**Categories:** Programming, Go
Short description...
**Comments:** 5 | **Likes:** 10 | **Dislikes:** 2
```

#### Post Page
```
## Post Title
**Author:** User1 | **Date:** 01.01.2023
**Categories:** Programming, Go

Full post content...

[Like] [Dislike] (10 likes, 2 dislikes)

### Comments
- **User2:** Interesting post! [Like] [Dislike] (3 likes)
- **User3:** Thanks for the information. [Like] [Dislike] (1 like)

[Add Comment] (form for authenticated users)
```

## 🚀 Installation

### Prerequisites
- Go 1.19 or higher
- Docker (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd forum
   ```

2. **Initialize Go module**
   ```bash
   go mod init forum
   ```

3. **Install dependencies**
   ```bash
   go get github.com/mattn/go-sqlite3
   go get github.com/google/uuid
   go get golang.org/x/crypto/bcrypt
   ```

4. **Set up the database**
   ```bash
   # Database will be created automatically on first run
   ```

5. **Run the application**
   ```bash
   go run main.go
   ```

6. **Access the forum**
   Open your browser and navigate to `http://localhost:8080`

### Docker Deployment

1. **Build the Docker image**
   ```bash
   docker build -t go-forum .
   ```

2. **Run the container**
   ```bash
   docker run -p 8080:8080 go-forum
   ```
3. **Build and run**
  ```bash
  docker compose up -d --build
   ```
4. **View logs**
  ```bash
  docker compose logs -f
   ```
5. **Stop the application**
  ```bash
  docker compose down
   ```
## 📋 Key Files Description

### Backend Structure
- **`main.go`** - Main application entry point and server setup
- **`error.go`** - Global error handling functions
- **`database/db_setup.go`** - Database initialization and setup
- **`database/forum.db`** - SQLite database file

### Handlers (Request Processing)
- **`handlers/registration.go`** - User registration logic
- **`handlers/login.go`** - User authentication
- **`handlers/logout.go`** - Session termination
- **`handlers/create-post.go`** - Post creation functionality
- **`handlers/comments.go`** - Comment management
- **`handlers/get-post.go`** - Individual post retrieval
- **`handlers/get-posts-list.go`** - Posts listing with pagination
- **`handlers/create-reactions.go`** - Like/dislike functionality
- **`handlers/get-reactions.go`** - Reaction data retrieval
- **`handlers/get-likes-count.go`** - Reaction counting
- **`handlers/get-categories.go`** - Category management
- **`handlers/profile.go`** - User profile management
- **`handlers/user.go`** - User-related operations

### Utilities
- **`utils/open_db.go`** - Database connection utilities
- **`utils/category.go`** - Category-related helper functions

### Frontend Templates
- **Layout & Navigation**
  - `layout.html` - Main page layout template
  - `header.html` - Page header component
  - `nav.html` - Navigation menu component

- **Authentication Pages**
  - `login.html` - Login form
  - `registration.html` - User registration form
  - `success.html` - Success confirmation page

- **Content Pages**
  - `index.html` - Homepage with posts feed
  - `post_page.html` - Individual post view
  - `post_list.html` - Posts listing page
  - `post_item.html` - Single post component
  - `post_list_item.html` - Post preview component
  - `create_post.html` - Post creation form
  - `view_posts.html` - Posts viewing interface

- **User & Comments**
  - `profile.html` - User profile page
  - `user_page.html` - Public user page
  - `comment.html` - Comment display component
  - `add_comment.html` - Comment creation form

- **Filtering & Categories**
  - `filters.html` - Content filtering interface

- **Error Pages**
  - `400.html` - Bad Request
  - `401.html` - Unauthorized Access
  - `404.html` - Page Not Found
  - `405.html` - Method Not Allowed
  - `409.html` - Conflict Error
  - `500.html` - Internal Server Error

### Configuration Files
- **`categories.txt`** - Available forum categories
- **`design-ua.txt`** - Ukrainian design specifications
- **`design-uk.txt`** - UK design specifications
- **`project_description.md`** - Detailed project description
- **`test.txt`** - Testing notes and data

### User Registration
1. Navigate to `/register`
2. Fill in username, email, and password
3. Submit the form to create your account

### Login
1. Navigate to `/login`
2. Enter your email and password
3. Access all forum features after successful authentication

### Creating Posts
1. Click "Create Post" (requires authentication)
2. Enter title and content
3. Select relevant categories
4. Publish your post

### Interacting with Content
- **Like/Dislike**: Click the respective buttons on posts and comments
- **Comment**: Add comments to posts you find interesting
- **Filter**: Use category filters to find specific content

## 📋 Development Plan

### Phase 1: Project Setup
- [x] Initialize Go project structure
- [x] Set up SQLite database
- [x] Create ER diagram
- [x] Install dependencies

### Phase 2: Authentication System
- [ ] User registration with validation
- [ ] Login with session management
- [ ] Password hashing with bcrypt
- [ ] Logout functionality

### Phase 3: Core Features
- [ ] Post creation and display
- [ ] Comment system
- [ ] Category management
- [ ] Like/Dislike system

### Phase 4: Advanced Features
- [ ] Post filtering by categories
- [ ] Filter by user's posts
- [ ] Filter by liked posts
- [ ] Search functionality

### Phase 5: Docker & Testing
- [ ] Create Dockerfile
- [ ] Write unit tests
- [ ] Integration testing
- [ ] Documentation

### Phase 6: Polish & Deployment
- [ ] UI/UX improvements
- [ ] Performance optimization
- [ ] Security audit
- [ ] Production deployment

## 🔌 API Endpoints

### Authentication
- `GET /register` - Registration form (`registration.html`)
- `POST /register` - Process registration (`registration.go`)
- `GET /login` - Login form (`login.html`)
- `POST /login` - Process login (`login.go`)
- `POST /logout` - Logout user (`logout.go`)

### Posts
- `GET /` - Home page with posts list (`index.html`)
- `GET /post/:id` - Individual post view (`post_page.html`)
- `GET /create-post` - Create post form (`create_post.html`)
- `POST /create-post` - Process post creation (`create-post.go`)
- `GET /posts` - View posts list (`view_posts.html`)

### Comments
- `POST /comment` - Add comment to post (`comments.go`)
- `GET /add-comment` - Add comment form (`add_comment.html`)

### Reactions (Likes/Dislikes)
- `POST /create-reaction` - Create like/dislike (`create-reactions.go`)
- `GET /reactions` - Get reactions for content (`get-reactions.go`)
- `GET /likes-count` - Get likes count (`get-likes-count.go`)

### Categories & Filtering
- `GET /categories` - Get all categories (`get-categories.go`)
- `GET /filters` - Filter interface (`filters.html`)

### User Management
- `GET /profile` - User profile page (`profile.html`)
- `GET /user/:id` - User page (`user_page.html`)
- `GET /user` - User operations (`user.go`)

### Error Pages
- `GET /400` - Bad Request (`400.html`)
- `GET /401` - Unauthorized (`401.html`)
- `GET /404` - Not Found (`404.html`)
- `GET /405` - Method Not Allowed (`405.html`)
- `GET /409` - Conflict (`409.html`)
- `GET /500` - Internal Server Error (`500.html`)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow Go best practices and conventions
- Write comprehensive tests for new features
- Update documentation for any API changes
- Ensure all tests pass before submitting PR

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📧 Support

For support and questions, please open an issue in the repository or contact the development team.

---

**Happy Coding! 🚀**