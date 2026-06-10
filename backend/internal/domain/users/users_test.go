package users_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go-box/backend/internal/domain/users"
)

// mockUserRepository is a helper mock for TDD.
type mockUserRepository struct {
	GetOrCreateFunc func(ctx context.Context, id string) (*users.User, error)
	UpdateQuotaFunc func(ctx context.Context, id string, bytesDelta int64) (*users.User, error)
}

func (m *mockUserRepository) GetOrCreate(ctx context.Context, id string) (*users.User, error) {
	if m.GetOrCreateFunc != nil {
		return m.GetOrCreateFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepository) UpdateQuota(ctx context.Context, id string, bytesDelta int64) (*users.User, error) {
	if m.UpdateQuotaFunc != nil {
		return m.UpdateQuotaFunc(ctx, id, bytesDelta)
	}
	return nil, nil
}

var _ = Describe("QuotaService", func() {
	var (
		repo         *mockUserRepository
		quotaService *users.QuotaService
		ctx          context.Context
	)

	BeforeEach(func() {
		repo = &mockUserRepository{}
		quotaService = users.NewQuotaService(repo)
		ctx = context.Background()
	})

	Describe("ValidateAndReserve", func() {
		Context("when the user ID is invalid", func() {
			It("should return ErrInvalidUserID if user ID is empty", func() {
				err := quotaService.ValidateAndReserve(ctx, "", 100)
				Expect(err).To(MatchError(users.ErrInvalidUserID))
			})
		})

		Context("when the user exists and has enough quota", func() {
			It("should successfully update/reserve the quota", func() {
				userID := "user-123"
				fileSize := int64(1024 * 1024) // 1 MB

				repo.GetOrCreateFunc = func(ctx context.Context, id string) (*users.User, error) {
					Expect(id).To(Equal(userID))
					return &users.User{ID: userID, UsedQuota: 0}, nil
				}

				repo.UpdateQuotaFunc = func(ctx context.Context, id string, delta int64) (*users.User, error) {
					Expect(id).To(Equal(userID))
					Expect(delta).To(Equal(fileSize))
					return &users.User{ID: userID, UsedQuota: fileSize}, nil
				}

				err := quotaService.ValidateAndReserve(ctx, userID, fileSize)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the user quota is exceeded", func() {
			It("should return ErrQuotaExceeded and not update the quota", func() {
				userID := "user-123"
				// 2 GB limit, requesting slightly more than 2 GB
				fileSize := users.MaxStorageQuota + 1

				repo.GetOrCreateFunc = func(ctx context.Context, id string) (*users.User, error) {
					return &users.User{ID: userID, UsedQuota: 0}, nil
				}

				calledUpdate := false
				repo.UpdateQuotaFunc = func(ctx context.Context, id string, delta int64) (*users.User, error) {
					calledUpdate = true
					return nil, nil
				}

				err := quotaService.ValidateAndReserve(ctx, userID, fileSize)
				Expect(err).To(MatchError(users.ErrQuotaExceeded))
				Expect(calledUpdate).To(BeFalse())
			})

			It("should account for existing used quota", func() {
				userID := "user-123"
				existingUsage := users.MaxStorageQuota - 100
				fileSize := int64(101) // exceeds by 1 byte

				repo.GetOrCreateFunc = func(ctx context.Context, id string) (*users.User, error) {
					return &users.User{ID: userID, UsedQuota: existingUsage}, nil
				}

				calledUpdate := false
				repo.UpdateQuotaFunc = func(ctx context.Context, id string, delta int64) (*users.User, error) {
					calledUpdate = true
					return nil, nil
				}

				err := quotaService.ValidateAndReserve(ctx, userID, fileSize)
				Expect(err).To(MatchError(users.ErrQuotaExceeded))
				Expect(calledUpdate).To(BeFalse())
			})
		})

		Context("when database errors occur", func() {
			It("should forward errors from GetOrCreate", func() {
				userID := "user-123"
				dbErr := errors.New("database connection failed")

				repo.GetOrCreateFunc = func(ctx context.Context, id string) (*users.User, error) {
					return nil, dbErr
				}

				err := quotaService.ValidateAndReserve(ctx, userID, 100)
				Expect(err).To(MatchError(dbErr))
			})

			It("should forward errors from UpdateQuota", func() {
				userID := "user-123"
				dbErr := errors.New("update conflict")

				repo.GetOrCreateFunc = func(ctx context.Context, id string) (*users.User, error) {
					return &users.User{ID: userID, UsedQuota: 0}, nil
				}

				repo.UpdateQuotaFunc = func(ctx context.Context, id string, delta int64) (*users.User, error) {
					return nil, dbErr
				}

				err := quotaService.ValidateAndReserve(ctx, userID, 100)
				Expect(err).To(MatchError(dbErr))
			})
		})
	})
})

