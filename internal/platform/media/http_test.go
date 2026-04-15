package media

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileHandlerServesStoredFile(t *testing.T) {
	dir := t.TempDir()
	filename := "avatar.png"
	content := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
		0x54, 0x08, 0x99, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92,
		0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(filepath.Join(dir, filename), content, 0o644); err != nil {
		t.Fatalf("write stored file: %v", err)
	}

	handler := NewLocalFileHandler(dir, "/uploads/avatars")
	req := httptest.NewRequest(http.MethodGet, "/uploads/avatars/"+filename, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.Bytes(); string(got) != string(content) {
		t.Fatalf("unexpected body length = %d, want %d", len(got), len(content))
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "image/png" {
		t.Fatalf("content-type = %q, want image/png", contentType)
	}
}

func TestLocalFileHandlerRejectsTraversal(t *testing.T) {
	handler := NewLocalFileHandler(t.TempDir(), "/uploads/events")
	req := httptest.NewRequest(http.MethodGet, "/uploads/../secret.txt", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
