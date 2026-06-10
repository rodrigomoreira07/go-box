package users

import (
	"context"
	"errors"
)

// MaxStorageQuota defines the limit of 2 GB per user.
const MaxStorageQuota int64 = 2 * 1024 * 1024 * 1024 // 2 GB

var (
	ErrQuotaExceeded   = errors.New("storage quota exceeded")
	ErrInvalidUserID   = errors.New("invalid or empty user ID")
	ErrNotImplemented  = errors.New("not implemented")
)

// User represents a tenant in the system, whose authentication is handled by AWS Cognito.
type User struct {
	ID        string // Cognito sub (UUID)
	UsedQuota int64  // Current total storage used in bytes
}

// UserRepository defines the persistence actions for user quota tracking.
type UserRepository interface {
	GetOrCreate(ctx context.Context, id string) (*User, error)
	UpdateQuota(ctx context.Context, id string, bytesDelta int64) (*User, error)
}

// QuotaService orchestrates validation and reservations of user storage capacity.
type QuotaService struct {
	repo UserRepository
}

// NewQuotaService creates a new QuotaService with the provided repository.
func NewQuotaService(repo UserRepository) *QuotaService {
	return &QuotaService{repo: repo}
}

// ValidateAndReserve checks if the file size fits within the user's remaining quota,
// and conditionally reserves/updates the quota if it does.
func (s *QuotaService) ValidateAndReserve(ctx context.Context, userID string, fileSize int64) error {
	// Stub implementation to allow tests to compile but fail
	return ErrNotImplemented
}
