package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go-box/backend/internal/api"
	"go-box/backend/internal/files"
	"go-box/backend/internal/users"
)

type mockFileService struct {
	RequestUploadURLFunc   func(ctx context.Context, userID string, path string, size int64) (string, error)
	RequestDownloadURLFunc func(ctx context.Context, userID string, path string) (string, error)
}

func (m *mockFileService) RequestUploadURL(ctx context.Context, userID string, path string, size int64) (string, error) {
	if m.RequestUploadURLFunc != nil {
		return m.RequestUploadURLFunc(ctx, userID, path, size)
	}
	return "", nil
}

func (m *mockFileService) RequestDownloadURL(ctx context.Context, userID string, path string) (string, error) {
	if m.RequestDownloadURLFunc != nil {
		return m.RequestDownloadURLFunc(ctx, userID, path)
	}
	return "", nil
}

var _ = Describe("API Handlers", func() {
	var (
		fileSvc *mockFileService
		handler *api.Handler
		mux     *http.ServeMux
	)

	BeforeEach(func() {
		fileSvc = &mockFileService{}
		handler = api.NewHandler(fileSvc)
		mux = http.NewServeMux()
		handler.RegisterRoutes(mux)
	})

	Describe("AuthMiddleware", func() {
		It("should return 401 Unauthorized if X-User-Id header is missing", func() {
			req := httptest.NewRequest("POST", "/api/files/upload-url", bytes.NewBufferString(`{}`))
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should return 401 Unauthorized if X-User-Id header is empty", func() {
			req := httptest.NewRequest("POST", "/api/files/upload-url", bytes.NewBufferString(`{}`))
			req.Header.Set("X-User-Id", "")
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("POST /api/files/upload-url", func() {
		It("should return 400 Bad Request if request body is not valid JSON", func() {
			req := httptest.NewRequest("POST", "/api/files/upload-url", bytes.NewBufferString(`{invalid-json`))
			req.Header.Set("X-User-Id", "user-123")
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 200 OK and S3 presigned URL on success", func() {
			expectedURL := "https://s3.amazonaws.com/upload"
			userID := "user-123"
			filePath := "documents/report.pdf"
			fileSize := int64(1024)

			fileSvc.RequestUploadURLFunc = func(ctx context.Context, uid string, path string, size int64) (string, error) {
				Expect(uid).To(Equal(userID))
				Expect(path).To(Equal(filePath))
				Expect(size).To(Equal(fileSize))
				return expectedURL, nil
			}

			body := map[string]interface{}{
				"path": filePath,
				"size": fileSize,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/files/upload-url", bytes.NewBuffer(bodyBytes))
			req.Header.Set("X-User-Id", userID)
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))

			var result map[string]string
			err := json.Unmarshal(resp.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["upload_url"]).To(Equal(expectedURL))
		})

		It("should map ErrQuotaExceeded to 413 Payload Too Large", func() {
			userID := "user-123"
			fileSvc.RequestUploadURLFunc = func(ctx context.Context, uid string, path string, size int64) (string, error) {
				return "", users.ErrQuotaExceeded
			}

			body := map[string]interface{}{
				"path": "test.txt",
				"size": 500,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/files/upload-url", bytes.NewBuffer(bodyBytes))
			req.Header.Set("X-User-Id", userID)
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusRequestEntityTooLarge))
		})

		It("should map ErrInvalidPath to 400 Bad Request", func() {
			userID := "user-123"
			fileSvc.RequestUploadURLFunc = func(ctx context.Context, uid string, path string, size int64) (string, error) {
				return "", files.ErrInvalidPath
			}

			body := map[string]interface{}{
				"path": "../traversal.txt",
				"size": 100,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/files/upload-url", bytes.NewBuffer(bodyBytes))
			req.Header.Set("X-User-Id", userID)
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("GET /api/files/download-url", func() {
		It("should return 400 Bad Request if path query parameter is empty", func() {
			req := httptest.NewRequest("GET", "/api/files/download-url", nil)
			req.Header.Set("X-User-Id", "user-123")
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 200 OK and presigned download URL on success", func() {
			expectedURL := "https://s3.amazonaws.com/download"
			userID := "user-123"
			filePath := "docs/report.pdf"

			fileSvc.RequestDownloadURLFunc = func(ctx context.Context, uid string, path string) (string, error) {
				Expect(uid).To(Equal(userID))
				Expect(path).To(Equal(filePath))
				return expectedURL, nil
			}

			req := httptest.NewRequest("GET", "/api/files/download-url?path=docs/report.pdf", nil)
			req.Header.Set("X-User-Id", userID)
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))

			var result map[string]string
			err := json.Unmarshal(resp.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["download_url"]).To(Equal(expectedURL))
		})

		It("should map ErrFileNotFound to 404 Not Found", func() {
			userID := "user-123"
			fileSvc.RequestDownloadURLFunc = func(ctx context.Context, uid string, path string) (string, error) {
				return "", files.ErrFileNotFound
			}

			req := httptest.NewRequest("GET", "/api/files/download-url?path=missing.txt", nil)
			req.Header.Set("X-User-Id", userID)
			resp := httptest.NewRecorder()

			mux.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})
	})
})
