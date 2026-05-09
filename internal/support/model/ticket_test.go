package model

import "testing"

func TestCategoryValid(t *testing.T) {
	for _, category := range []Category{CategoryBug, CategorySuggestion, CategoryProductComplaint, CategoryOther} {
		if !category.Valid() {
			t.Fatalf("category %q should be valid", category)
		}
	}
	if Category("billing").Valid() {
		t.Fatal("unknown category should be invalid")
	}
}

func TestStatusValid(t *testing.T) {
	for _, status := range []Status{StatusOpen, StatusInProgress, StatusClosed} {
		if !status.Valid() {
			t.Fatalf("status %q should be valid", status)
		}
	}
	if Status("archived").Valid() {
		t.Fatal("unknown status should be invalid")
	}
}

func TestMessageAuthorRoleValid(t *testing.T) {
	for _, role := range []MessageAuthorRole{MessageAuthorRoleUser, MessageAuthorRoleSupport, MessageAuthorRoleAdmin} {
		if !role.Valid() {
			t.Fatalf("message author role %q should be valid", role)
		}
	}
	if MessageAuthorRole("robot").Valid() {
		t.Fatal("unknown message author role should be invalid")
	}
}
