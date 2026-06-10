package files_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go-box/backend/internal/domain/files"
)

type mockFileRepository struct {
	SaveFunc       func(ctx context.Context, file *files.File) error
	DeleteFunc     func(ctx context.Context, userID string, path string) error
	FindByPathFunc func(ctx context.Context, userID string, path string) (*files.File, error)
	ListByUserFunc func(ctx context.Context, userID string) ([]*files.File, error)
}

func (m *mockFileRepository) Save(ctx context.Context, file *files.File) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, file)
	}
	return nil
}

func (m *mockFileRepository) Delete(ctx context.Context, userID string, path string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, userID, path)
	}
	return nil
}

func (m *mockFileRepository) FindByPath(ctx context.Context, userID string, path string) (*files.File, error) {
	if m.FindByPathFunc != nil {
		return m.FindByPathFunc(ctx, userID, path)
	}
	return nil, nil
}

func (m *mockFileRepository) ListByUser(ctx context.Context, userID string) ([]*files.File, error) {
	if m.ListByUserFunc != nil {
		return m.ListByUserFunc(ctx, userID)
	}
	return nil, nil
}

type mockStorageService struct {
	GenerateUploadURLFunc   func(ctx context.Context, userID string, path string) (string, error)
	GenerateDownloadURLFunc func(ctx context.Context, userID string, path string) (string, error)
}

func (m *mockStorageService) GenerateUploadURL(ctx context.Context, userID string, path string) (string, error) {
	if m.GenerateUploadURLFunc != nil {
		return m.GenerateUploadURLFunc(ctx, userID, path)
	}
	return "", nil
}

func (m *mockStorageService) GenerateDownloadURL(ctx context.Context, userID string, path string) (string, error) {
	if m.GenerateDownloadURLFunc != nil {
		return m.GenerateDownloadURLFunc(ctx, userID, path)
	}
	return "", nil
}

type mockQuotaValidator struct {
	ValidateAndReserveFunc func(ctx context.Context, userID string, fileSize int64) error
}

func (m *mockQuotaValidator) ValidateAndReserve(ctx context.Context, userID string, fileSize int64) error {
	if m.ValidateAndReserveFunc != nil {
		return m.ValidateAndReserveFunc(ctx, userID, fileSize)
	}
	return nil
}

var _ = Describe("FileService", func() {
	var (
		repo         *mockFileRepository
		storage      *mockStorageService
		quotaVal     *mockQuotaValidator
		fileService  *files.FileService
		ctx          context.Context
		userID       string
	)

	BeforeEach(func() {
		repo = &mockFileRepository{}
		storage = &mockStorageService{}
		quotaVal = &mockQuotaValidator{}
		fileService = files.NewFileService(repo, storage, quotaVal)
		ctx = context.Background()
		userID = "cognito-uuid-12345"
	})

	Describe("RequestUploadURL", func() {
		Context("when parameters are invalid", func() {
			It("should return ErrInvalidUserID if userID is empty", func() {
				_, err := fileService.RequestUploadURL(ctx, "", "file.txt", 100)
				Expect(err).To(MatchError(files.ErrInvalidUserID))
			})

			It("should return ErrInvalidPath on path traversal attempts", func() {
				_, err := fileService.RequestUploadURL(ctx, userID, "../traversal.txt", 100)
				Expect(err).To(MatchError(files.ErrInvalidPath))

				_, err = fileService.RequestUploadURL(ctx, userID, "some/../../nested.txt", 100)
				Expect(err).To(MatchError(files.ErrInvalidPath))
			})

			It("should return ErrDirectoryUpload if path suggests directory upload", func() {
				_, err := fileService.RequestUploadURL(ctx, userID, "photos/", 0)
				Expect(err).To(MatchError(files.ErrDirectoryUpload))
			})
		})

		Context("when quota check fails", func() {
			It("should forward quota error and not generate URL", func() {
				quotaErr := errors.New("quota limit reached")
				quotaVal.ValidateAndReserveFunc = func(ctx context.Context, uid string, size int64) error {
					Expect(uid).To(Equal(userID))
					Expect(size).To(Equal(int64(1024)))
					return quotaErr
				}

				calledStorage := false
				storage.GenerateUploadURLFunc = func(ctx context.Context, uid string, path string) (string, error) {
					calledStorage = true
					return "", nil
				}

				_, err := fileService.RequestUploadURL(ctx, userID, "folder/file.txt", 1024)
				Expect(err).To(MatchError(quotaErr))
				Expect(calledStorage).To(BeFalse())
			})
		})

		Context("when parameters, quota, storage, and database save succeed", func() {
			It("should return the S3 URL and save file metadata", func() {
				expectedURL := "https://s3.amazonaws.com/gobox-bucket/users/cognito-uuid-12345/docs/report.pdf?signature=123"
				filePath := "docs/report.pdf"
				fileSize := int64(2048)

				quotaVal.ValidateAndReserveFunc = func(ctx context.Context, uid string, size int64) error {
					return nil
				}

				storage.GenerateUploadURLFunc = func(ctx context.Context, uid string, path string) (string, error) {
					Expect(uid).To(Equal(userID))
					Expect(path).To(Equal(filePath))
					return expectedURL, nil
				}

				calledSave := false
				repo.SaveFunc = func(ctx context.Context, file *files.File) error {
					Expect(file.UserID).To(Equal(userID))
					Expect(file.Path).To(Equal(filePath))
					Expect(file.Size).To(Equal(fileSize))
					Expect(file.IsDir).To(BeFalse())
					calledSave = true
					return nil
				}

				url, err := fileService.RequestUploadURL(ctx, userID, filePath, fileSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(url).To(Equal(expectedURL))
				Expect(calledSave).To(BeTrue())
			})
		})
	})

	Describe("RequestDownloadURL", func() {
		Context("when parameters are invalid", func() {
			It("should return ErrInvalidUserID if userID is empty", func() {
				_, err := fileService.RequestDownloadURL(ctx, "", "file.txt")
				Expect(err).To(MatchError(files.ErrInvalidUserID))
			})

			It("should return ErrInvalidPath on path traversal attempts", func() {
				_, err := fileService.RequestDownloadURL(ctx, userID, "../traversal.txt")
				Expect(err).To(MatchError(files.ErrInvalidPath))
			})
		})

		Context("when user has no access to the file (Multi-Tenancy Guard)", func() {
			It("should return ErrFileNotFound if record belongs to another tenant", func() {
				filePath := "confidential.pdf"
				repo.FindByPathFunc = func(ctx context.Context, uid string, path string) (*files.File, error) {
					// In a multi-tenant DB query, we query by path and userID.
					// If it doesn't match, the DB returns nil/no rows.
					return nil, nil
				}

				_, err := fileService.RequestDownloadURL(ctx, userID, filePath)
				Expect(err).To(MatchError(files.ErrFileNotFound))
			})
		})

		Context("when user owns the file metadata", func() {
			It("should return presigned download URL", func() {
				filePath := "my-file.txt"
				expectedURL := "https://s3.amazonaws.com/gobox-bucket/users/cognito-uuid-12345/my-file.txt?signature=987"

				repo.FindByPathFunc = func(ctx context.Context, uid string, path string) (*files.File, error) {
					Expect(uid).To(Equal(userID))
					Expect(path).To(Equal(filePath))
					return &files.File{
						ID:     "file-id-1",
						UserID: userID,
						Path:   filePath,
						Size:   500,
						IsDir:  false,
					}, nil
				}

				storage.GenerateDownloadURLFunc = func(ctx context.Context, uid string, path string) (string, error) {
					Expect(uid).To(Equal(userID))
					Expect(path).To(Equal(filePath))
					return expectedURL, nil
				}

				url, err := fileService.RequestDownloadURL(ctx, userID, filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(url).To(Equal(expectedURL))
			})
		})
	})
})
