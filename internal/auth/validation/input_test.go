package validation

import "testing"

func TestValidateRegisterSuccess(t *testing.T) {
	email, password, err := ValidateRegister(
		" Tester@Example.com ",
		"verysecret",
	)
	if err != nil {
		t.Fatalf("ValidateRegister() error = %v", err)
	}
	if email != "tester@example.com" {
		t.Fatalf("email = %q, want normalized value", email)
	}
	if password != "verysecret" {
		t.Fatalf("password = %q, want normalized value", password)
	}
}

func TestValidateRegisterValidationError(t *testing.T) {
	_, _, err := ValidateRegister("", "123")
	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("error type = %T, want ValidationError", err)
	}
	for _, key := range []string{"email", "password"} {
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
	surname := " Петров "
	birthday := "2001-02-03"
	cityID := "AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA"
	avatarURL := "/uploads/avatars/file.png"

	gotEmail, gotUsername, gotSurname, gotBirthday, gotCityID, gotAvatarURL, _, _, err := ValidateProfilePatch(
		&email, &username, &surname, &birthday, &cityID, &avatarURL, nil, nil,
	)
	if err != nil {
		t.Fatalf("ValidateProfilePatch() error = %v", err)
	}
	if gotEmail != "user@example.com" {
		t.Fatalf("email = %q, want normalized value", gotEmail)
	}
	if gotUsername != "user-1" || gotSurname != "Петров" {
		t.Fatalf("unexpected normalized names: %q %q", gotUsername, gotSurname)
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
	surname := ""
	birthday := "bad"
	cityID := "bad"
	avatarURL := "file.png"

	_, _, _, _, _, _, _, _, err := ValidateProfilePatch(&email, &username, &surname, &birthday, &cityID, &avatarURL, nil, nil)
	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("error type = %T, want ValidationError", err)
	}
	for _, key := range []string{"email", "username", "userSurname", "birthday", "cityId", "avatarUrl"} {
		if _, exists := validationErr.Details[key]; !exists {
			t.Fatalf("missing validation detail for %q: %+v", key, validationErr.Details)
		}
	}
}
