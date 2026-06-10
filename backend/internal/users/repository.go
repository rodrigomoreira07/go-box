package users

import "context"

// Repository defines the persistence actions for user quota tracking.
type Repository interface {
	Get(ctx context.Context, id string) (*User, error)
	UpdateQuota(ctx context.Context, id string, bytesDelta int64) (*User, error)
}
