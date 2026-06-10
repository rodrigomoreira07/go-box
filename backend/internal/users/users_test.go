package users_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go-box/backend/internal/users"
)

// mockUserRepository is a helper mock for TDD.
type mockUserRepository struct {
	GetFunc         func(ctx context.Context, id string) (*users.User, error)
	UpdateQuotaFunc func(ctx context.Context, id string, bytesDelta int64) (*users.User, error)
}

func (m *mockUserRepository) Get(ctx context.Context, id string) (*users.User, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
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

	Describe("CheckQuota", func() {
		Context("when the user ID is invalid", func() {
			It("should return ErrInvalidUserID if user ID is empty", func() {
				err := quotaService.CheckQuota(ctx, "", 100)
				Expect(err).To(MatchError(users.ErrInvalidUserID))
			})
		})

		Context("when the user exists and has enough quota", func() {
			It("should return nil", func() {
				userID := "user-123"
				fileSize := int64(1024 * 1024) // 1 MB

				repo.GetFunc = func(ctx context.Context, id string) (*users.User, error) {
					Expect(id).To(Equal(userID))
					return &users.User{
						ID:        userID,
						UsedQuota: 0,
						MaxQuota:  users.MaxStorageQuota,
					}, nil
				}

				err := quotaService.CheckQuota(ctx, userID, fileSize)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the user quota is exceeded", func() {
			It("should return ErrQuotaExceeded", func() {
				userID := "user-123"
				fileSize := users.MaxStorageQuota + 1

				repo.GetFunc = func(ctx context.Context, id string) (*users.User, error) {
					return &users.User{
						ID:        userID,
						UsedQuota: 0,
						MaxQuota:  users.MaxStorageQuota,
					}, nil
				}

				err := quotaService.CheckQuota(ctx, userID, fileSize)
				Expect(err).To(MatchError(users.ErrQuotaExceeded))
			})

			It("should account for existing used quota", func() {
				userID := "user-123"
				existingUsage := users.MaxStorageQuota - 100
				fileSize := int64(101) // exceeds by 1 byte

				repo.GetFunc = func(ctx context.Context, id string) (*users.User, error) {
					return &users.User{
						ID:        userID,
						UsedQuota: existingUsage,
						MaxQuota:  users.MaxStorageQuota,
					}, nil
				}

				err := quotaService.CheckQuota(ctx, userID, fileSize)
				Expect(err).To(MatchError(users.ErrQuotaExceeded))
			})
		})

		Context("when database errors occur", func() {
			It("should forward errors from Get", func() {
				userID := "user-123"
				dbErr := errors.New("database connection failed")

				repo.GetFunc = func(ctx context.Context, id string) (*users.User, error) {
					return nil, dbErr
				}

				err := quotaService.CheckQuota(ctx, userID, 100)
				Expect(err).To(MatchError(dbErr))
			})
		})
	})

	Describe("ConsumeQuota", func() {
		Context("when the user ID is invalid", func() {
			It("should return ErrInvalidUserID if user ID is empty", func() {
				err := quotaService.ConsumeQuota(ctx, "", 100)
				Expect(err).To(MatchError(users.ErrInvalidUserID))
			})
		})

		Context("when the database succeeds", func() {
			It("should call UpdateQuota and return nil", func() {
				userID := "user-123"
				fileSize := int64(100)

				calledUpdate := false
				repo.UpdateQuotaFunc = func(ctx context.Context, id string, delta int64) (*users.User, error) {
					Expect(id).To(Equal(userID))
					Expect(delta).To(Equal(fileSize))
					calledUpdate = true
					return &users.User{ID: userID, UsedQuota: fileSize}, nil
				}

				err := quotaService.ConsumeQuota(ctx, userID, fileSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(calledUpdate).To(BeTrue())
			})
		})

		Context("when database errors occur", func() {
			It("should forward errors from UpdateQuota", func() {
				userID := "user-123"
				dbErr := errors.New("update conflict")

				repo.UpdateQuotaFunc = func(ctx context.Context, id string, delta int64) (*users.User, error) {
					return nil, dbErr
				}

				err := quotaService.ConsumeQuota(ctx, userID, 100)
				Expect(err).To(MatchError(dbErr))
			})
		})
	})
})
