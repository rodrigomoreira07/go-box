package files

import (
	"errors"
	"time"
)

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrFileNotFound    = errors.New("file not found")
	ErrInvalidPath     = errors.New("invalid file path or traversal attempt")
	ErrDirectoryUpload = errors.New("cannot upload directly to a directory")
	ErrInvalidUserID   = errors.New("invalid or empty user ID")
)

// File represents a file or folder metadata record.
type File struct {
	ID        string
	UserID    string // Cognito sub
	Path      string // Relative path, e.g. "documents/invoice.pdf"
	Size      int64  // Size in bytes (0 for directories)
	SHA256    string // SHA-256 hash (empty for directories)
	IsDir     bool   // Indicates if metadata represents a directory
	CreatedAt time.Time
	UpdatedAt time.Time
}
