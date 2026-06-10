package files

import "context"

// Repository defines database access patterns for file metadata.
type Repository interface {
	Save(ctx context.Context, file *File) error
	Delete(ctx context.Context, userID string, path string) error
	FindByPath(ctx context.Context, userID string, path string) (*File, error)
	ListByUser(ctx context.Context, userID string) ([]*File, error)
}
