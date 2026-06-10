package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go-box/backend/internal/files"
	"go-box/backend/internal/users"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// FileService defines the orchestration methods needed by HTTP handlers.
type FileService interface {
	RequestUploadURL(ctx context.Context, userID string, path string, size int64) (string, error)
	RequestDownloadURL(ctx context.Context, userID string, path string) (string, error)
}

// Handler handles HTTP requests for GoBox backend features.
type Handler struct {
	fileSvc FileService
}

// NewHandler creates a new Handler instance.
func NewHandler(fileSvc FileService) *Handler {
	return &Handler{fileSvc: fileSvc}
}

// AuthMiddleware extracts the X-User-Id Cognito header and places it in the context.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-Id")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RegisterRoutes registers the REST endpoints with the multiplexer.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/files/upload-url", h.AuthMiddleware(http.HandlerFunc(h.handleUploadURL)))
	mux.Handle("GET /api/files/download-url", h.AuthMiddleware(http.HandlerFunc(h.handleDownloadURL)))
}

func (h *Handler) handleUploadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	url, err := h.fileSvc.RequestUploadURL(r.Context(), userID, body.Path, body.Size)
	if err != nil {
		if errors.Is(err, users.ErrQuotaExceeded) {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		if errors.Is(err, files.ErrInvalidPath) || errors.Is(err, files.ErrDirectoryUpload) || errors.Is(err, files.ErrInvalidUserID) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"upload_url": url})
}

func (h *Handler) handleDownloadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}

	url, err := h.fileSvc.RequestDownloadURL(r.Context(), userID, path)
	if err != nil {
		if errors.Is(err, files.ErrFileNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, files.ErrInvalidPath) || errors.Is(err, files.ErrInvalidUserID) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"download_url": url})
}
