package media

import (
	"path"
	"strings"
)

func IsHTTPURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func IsUploadPath(value string) bool {
	if !strings.HasPrefix(value, "/uploads/") {
		return false
	}

	cleaned := path.Clean(value)
	return cleaned == value && cleaned != "/uploads" && !strings.Contains(value, `\`)
}

func IsFileReference(value string) bool {
	return IsHTTPURL(value) || IsUploadPath(value)
}
