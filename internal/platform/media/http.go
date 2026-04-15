package media

import (
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type LocalFileHandler struct {
	dir    string
	prefix string
}

func NewLocalFileHandler(dir, prefix string) *LocalFileHandler {
	return &LocalFileHandler{
		dir:    dir,
		prefix: strings.TrimRight(prefix, "/"),
	}
}

func (h *LocalFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.URL.Path, h.prefix+"/") {
		http.NotFound(w, r)
		return
	}

	publicPath := r.URL.Path
	if !IsUploadPath(publicPath) {
		http.NotFound(w, r)
		return
	}

	relativePath := strings.TrimPrefix(publicPath, h.prefix+"/")
	if relativePath == "" {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join(h.dir, filepath.FromSlash(path.Clean("/"+relativePath)))
	rel, err := filepath.Rel(h.dir, filePath)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		http.NotFound(w, r)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(header[:n]))
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodHead {
		return
	}

	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}
}
