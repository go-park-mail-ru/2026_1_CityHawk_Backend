package errors

import (
	"errors"
	"testing"
)

func TestInvalidReferenceError(t *testing.T) {
	err := NewInvalidReferenceError("authorId", "author not found", "event_author_id_fkey")
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("error does not unwrap to ErrInvalidReference: %v", err)
	}
	if got := err.Error(); got != "invalid reference: author not found" {
		t.Fatalf("Error() = %q", got)
	}

	field, message, ok := InvalidReferenceDetails(err)
	if !ok || field != "authorId" || message != "author not found" {
		t.Fatalf("InvalidReferenceDetails() = %q, %q, %v", field, message, ok)
	}

	if _, _, ok := InvalidReferenceDetails(ErrInvalidReference); ok {
		t.Fatal("plain ErrInvalidReference unexpectedly has details")
	}
}

func TestNilInvalidReferenceErrorMessage(t *testing.T) {
	var err *InvalidReferenceError
	if got := err.Error(); got != ErrInvalidReference.Error() {
		t.Fatalf("nil Error() = %q", got)
	}

	err = &InvalidReferenceError{}
	if got := err.Error(); got != ErrInvalidReference.Error() {
		t.Fatalf("empty Error() = %q", got)
	}
}
