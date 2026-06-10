package users

import "errors"

// MaxStorageQuota defines the limit of 2 GB per user.
const MaxStorageQuota int64 = 2 * 1024 * 1024 * 1024 // 2 GB

var (
	ErrQuotaExceeded  = errors.New("storage quota exceeded")
	ErrInvalidUserID  = errors.New("invalid or empty user ID")
	ErrNotImplemented = errors.New("not implemented")
)

// User represents a tenant in the system, whose authentication is handled by AWS Cognito.
type User struct {
	ID        string // Cognito sub (UUID)
	UsedQuota int64  // Current total storage used in bytes
	MaxQuota  int64  // Maximum storage limit allowed in bytes
}

// HasSufficientQuota checks if the requested fileSize fits within the user's remaining quota limit.
func (u *User) HasSufficientQuota(fileSize int64) bool {
	return u.UsedQuota+fileSize <= u.MaxQuota
}
