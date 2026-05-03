package model

import "testing"

func TestRoleValid(t *testing.T) {
	for _, role := range []Role{RoleUser, RoleOrganizer, RoleAdmin} {
		if !role.Valid() {
			t.Fatalf("role %q should be valid", role)
		}
	}
	if Role("moderator").Valid() {
		t.Fatal("unknown role should be invalid")
	}
}
