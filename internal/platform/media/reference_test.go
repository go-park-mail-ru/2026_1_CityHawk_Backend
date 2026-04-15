package media

import "testing"

func TestReferencePredicates(t *testing.T) {
	t.Run("http url", func(t *testing.T) {
		if !IsHTTPURL("https://example.com/image.jpg") {
			t.Fatal("IsHTTPURL() = false, want true")
		}
		if IsHTTPURL("/uploads/image.jpg") {
			t.Fatal("IsHTTPURL() = true for upload path, want false")
		}
	})

	t.Run("upload path", func(t *testing.T) {
		if !IsUploadPath("/uploads/image.jpg") {
			t.Fatal("IsUploadPath() = false, want true")
		}
		if IsUploadPath("/uploads/../etc/passwd") {
			t.Fatal("IsUploadPath() = true for cleaned traversal path, want false")
		}
		if IsUploadPath(`/uploads\image.jpg`) {
			t.Fatal("IsUploadPath() = true for windows-style path, want false")
		}
	})

	t.Run("file reference", func(t *testing.T) {
		if !IsFileReference("http://example.com/image.jpg") {
			t.Fatal("IsFileReference() = false for http URL, want true")
		}
		if !IsFileReference("/uploads/image.jpg") {
			t.Fatal("IsFileReference() = false for upload path, want true")
		}
		if IsFileReference("image.jpg") {
			t.Fatal("IsFileReference() = true for plain file name, want false")
		}
	})
}
