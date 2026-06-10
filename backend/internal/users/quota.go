package users

import "context"

// QuotaService orchestrates validation and reservations of user storage capacity.
type QuotaService struct {
	repo Repository
}

// NewQuotaService creates a new QuotaService with the provided repository.
func NewQuotaService(repo Repository) *QuotaService {
	return &QuotaService{repo: repo}
}

// CheckQuota verifies if the user has enough quota for the requested file size.
func (s *QuotaService) CheckQuota(ctx context.Context, userID string, fileSize int64) error {
	if userID == "" {
		return ErrInvalidUserID
	}

	u, err := s.repo.Get(ctx, userID)
	if err != nil {
		return err
	}

	if !u.HasSufficientQuota(fileSize) {
		return ErrQuotaExceeded
	}

	return nil
}

// ConsumeQuota updates/reserves the user's used storage space.
func (s *QuotaService) ConsumeQuota(ctx context.Context, userID string, fileSize int64) error {
	if userID == "" {
		return ErrInvalidUserID
	}

	_, err := s.repo.UpdateQuota(ctx, userID, fileSize)
	if err != nil {
		return err
	}

	return nil
}
