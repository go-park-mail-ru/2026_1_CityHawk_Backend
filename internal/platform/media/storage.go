package media

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

const MaxImageSize int64 = 5 << 20

var (
	ErrImageEmpty           = errors.New("image file is empty")
	ErrImageTooLarge        = errors.New("image file is too large")
	ErrUnsupportedImageType = errors.New("image must be a PNG, JPEG, GIF, or WebP file")
)

var imageExtensions = map[string]string{
	"image/gif":  ".gif",
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type Storage interface {
	Save(ownerID string, header *multipart.FileHeader) (string, error)
}

type LocalStorage struct {
	dir          string
	publicPrefix string
}

func NewLocalStorage(dir, publicPrefix string) (*LocalStorage, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	return &LocalStorage{
		dir:          dir,
		publicPrefix: publicPrefix,
	}, nil
}

func (s *LocalStorage) Dir() string {
	return s.dir
}

func (s *LocalStorage) Save(ownerID string, header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, MaxImageSize+1))
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", ErrImageEmpty
	}
	if int64(len(data)) > MaxImageSize {
		return "", ErrImageTooLarge
	}

	contentType := http.DetectContentType(data)
	ext, ok := imageExtensions[contentType]
	if !ok {
		return "", ErrUnsupportedImageType
	}

	filename := fmt.Sprintf("%s-%d%s", ownerID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(s.dir, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", err
	}

	return path.Join(s.publicPrefix, filename), nil
}
