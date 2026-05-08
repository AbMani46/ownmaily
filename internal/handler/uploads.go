package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxImageSize = 5 << 20 // 5 MB

var allowedMIMETypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// Storage persists an uploaded file and returns its public URL.
// LocalStorage writes to disk; a future implementation (R2, S3) implements the same interface.
type Storage interface {
	Save(filename string, r io.Reader) (url string, err error)
}

// LocalStorage writes uploads to a local directory and serves them via /uploads/images/.
// Save returns a root-relative URL (/uploads/images/<filename>) so the stored path
// remains portable when installationURL changes. Callers absolutize the URL at use time.
type LocalStorage struct {
	dir string
}

func NewLocalStorage(dir string) *LocalStorage {
	return &LocalStorage{dir: dir}
}

func (s *LocalStorage) Save(filename string, r io.Reader) (string, error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", fmt.Errorf("create upload dir: %w", err)
	}
	f, err := os.Create(filepath.Join(s.dir, filename))
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return "/uploads/images/" + filename, nil
}

type UploadHandler struct {
	storage Storage
}

func NewUploadHandler(storage Storage) *UploadHandler {
	return &UploadHandler{storage: storage}
}

func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large", "file exceeds 5 MB limit")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing_file", "missing file field")
		return
	}
	defer file.Close()

	// Sniff content type from actual bytes — do not trust the client-supplied MIME type.
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mime := http.DetectContentType(buf[:n])

	ext, ok := allowedMIMETypes[mime]
	if !ok {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_type", "accepted formats: jpeg, png, gif, webp")
		return
	}

	// Seek back so the full file (including the sniffed bytes) is written to storage.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeError(w, http.StatusInternalServerError, "seek_error", "internal error")
		return
	}

	filename := randomHex(16) + ext
	url, err := h.storage.Save(filename, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "upload_failed", "upload failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
