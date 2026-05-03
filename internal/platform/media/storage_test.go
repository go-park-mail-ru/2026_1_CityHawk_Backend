package media

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"testing"
)

func TestLocalStorageSave(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "/uploads/avatars")
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	publicPath, err := storage.Save("user-1", multipartHeader(t, "avatar.png", pngBytes()))
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if !strings.HasPrefix(publicPath, "/uploads/avatars/user-1-") || !strings.HasSuffix(publicPath, ".png") {
		t.Fatalf("public path = %q", publicPath)
	}
	if _, err := os.Stat(storage.Dir() + "/" + strings.TrimPrefix(publicPath, "/uploads/avatars/")); err != nil {
		t.Fatalf("saved file is missing: %v", err)
	}
}

func TestLocalStorageSaveRejectsBadImages(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "/uploads/events")
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "empty", data: nil, want: ErrImageEmpty},
		{name: "text", data: []byte("hello"), want: ErrUnsupportedImageType},
		{name: "too large", data: bytes.Repeat([]byte{0}, int(MaxImageSize)+1), want: ErrImageTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.Save("event-1", multipartHeader(t, "image.bin", tt.data))
			if !errors.Is(err, tt.want) {
				t.Fatalf("Save error = %v, want %v", err, tt.want)
			}
		})
	}
}

func multipartHeader(t *testing.T, filename string, data []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		t.Fatalf("write multipart data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(MaxImageSize + 1024)
	if err != nil {
		t.Fatalf("ReadForm: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("file count = %d, want 1", len(files))
	}
	return files[0]
}

func pngBytes() []byte {
	return []byte{
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
}
