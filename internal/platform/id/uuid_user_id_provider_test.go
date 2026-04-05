package id

import (
	"regexp"
	"testing"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestUUIDUserIDProviderNewReturnsUUIDv4(t *testing.T) {
	provider := NewUUIDUserIDProvider()

	id := provider.New()
	if !uuidPattern.MatchString(id) {
		t.Fatalf("New() = %q, want UUIDv4", id)
	}
}
