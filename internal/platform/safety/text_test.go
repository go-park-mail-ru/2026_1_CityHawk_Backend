package safety

import "testing"

func TestEscapeTextAndPointer(t *testing.T) {
	if got := EscapeText("<b>bold</b>"); got != "&lt;b&gt;bold&lt;/b&gt;" {
		t.Fatalf("EscapeText() = %q", got)
	}
	if EscapeTextPtr(nil) != nil {
		t.Fatal("EscapeTextPtr(nil) should return nil")
	}
	value := "<script>"
	escaped := EscapeTextPtr(&value)
	if escaped == nil || *escaped != "&lt;script&gt;" {
		t.Fatalf("EscapeTextPtr() = %#v", escaped)
	}
}
