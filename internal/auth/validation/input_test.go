package validation

import "testing"

func TestValidateRegisterSuccess(t *testing.T) {
	email, username, password, err := ValidateRegister(
		" Tester@Example.com ",
		" Алиса ",
		"verysecret",
	)
	if err != nil {
		t.Fatalf("ValidateRegister() error = %v", err)
	}
	if email != "tester@example.com" {
		t.Fatalf("email = %q, want normalized value", email)
	}
	if username != "Алиса" {
		t.Fatalf("username = %q, want normalized value", username)
	}
	if password != "verysecret" {
		t.Fatalf("password = %q, want normalized value", password)
	}
}

func TestValidateRegisterValidationError(t *testing.T) {
	_, _, _, err := ValidateRegister("", "", "123")
	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("error type = %T, want ValidationError", err)
	}
	for _, key := range []string{"email", "username", "password"} {
		if _, exists := validationErr.Details[key]; !exists {
			t.Fatalf("missing validation detail for %q: %+v", key, validationErr.Details)
		}
	}
}

func TestValidateLoginValidationError(t *testing.T) {
	_, _, err := ValidateLogin("bad", "")
	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("error type = %T, want ValidationError", err)
	}
	if _, exists := validationErr.Details["email"]; !exists {
		t.Fatalf("missing email error: %+v", validationErr.Details)
	}
	if _, exists := validationErr.Details["password"]; !exists {
		t.Fatalf("missing password error: %+v", validationErr.Details)
	}
}

func TestValidateProfilePatchAcceptsUploadPath(t *testing.T) {
	email := " USER@Example.COM "
	username := " user-1 "
	birthday := "2001-02-03"
	cityID := "AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA"
	avatarURL := "/uploads/avatars/file.png"

	gotEmail, gotUsername, gotBirthday, gotCityID, gotAvatarURL, _, _, err := ValidateProfilePatch(
		&email, &username, &birthday, &cityID, &avatarURL, nil, nil,
	)
	if err != nil {
		t.Fatalf("ValidateProfilePatch() error = %v", err)
	}
	if gotEmail != "user@example.com" {
		t.Fatalf("email = %q, want normalized value", gotEmail)
	}
	if gotUsername != "user-1" {
		t.Fatalf("username = %q, want normalized value", gotUsername)
	}
	if gotBirthday == nil || gotBirthday.Format("2006-01-02") != "2001-02-03" {
		t.Fatalf("birthday = %#v, want parsed date", gotBirthday)
	}
	if gotCityID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("cityID = %q, want lowercase uuid", gotCityID)
	}
	if gotAvatarURL != "/uploads/avatars/file.png" {
		t.Fatalf("avatarURL = %q", gotAvatarURL)
	}
}

func TestValidateProfilePatchValidationError(t *testing.T) {
	email := "bad"
	username := "!"
	birthday := "bad"
	cityID := "bad"
	avatarURL := "file.png"

	_, _, _, _, _, _, _, err := ValidateProfilePatch(&email, &username, &birthday, &cityID, &avatarURL, nil, nil)
	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("error type = %T, want ValidationError", err)
	}
	for _, key := range []string{"email", "username", "birthday", "cityId", "avatarUrl"} {
		if _, exists := validationErr.Details[key]; !exists {
			t.Fatalf("missing validation detail for %q: %+v", key, validationErr.Details)
		}
	}
}
