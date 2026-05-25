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

	t.Run("public url", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "https://static.cityhawk.ru/")
		if got := PublicURL("/uploads/avatars/image.jpg"); got != "https://static.cityhawk.ru/uploads/avatars/image.jpg" {
			t.Fatalf("PublicURL() = %q, want configured uploads URL", got)
		}
		if got := PublicURL("https://example.com/image.jpg"); got != "https://example.com/image.jpg" {
			t.Fatalf("PublicURL() changed external URL: %q", got)
		}
	})

	t.Run("public url default", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "")
		if got := PublicURL("/uploads/avatars/image.jpg"); got != "http://cityhawk.ru:8080/uploads/avatars/image.jpg" {
			t.Fatalf("PublicURL() = %q, want default cityhawk uploads URL", got)
		}
	})
}
