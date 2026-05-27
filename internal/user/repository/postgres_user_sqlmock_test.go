package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestScanUserWithSQLMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, time.April, 15, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "email", "username", "password_hash", "birthday", "city_id", "avatar_url", "bio", "role", "interest_tag_ids", "created_at", "updated_at",
		"city_record_id", "city_name", "country_name", "timezone",
	}).AddRow(
		"user-1", "user@example.com", "user_1", "hash", now, "city-1", "/uploads/avatars/file.png", "bio", "user", nil, now, now,
		"city-1", "Москва", "Россия", "Europe/Moscow",
	)
	mock.ExpectQuery("SELECT \\* FROM user_account").WillReturnRows(rows)

	row := db.QueryRow("SELECT * FROM user_account")
	user, err := scanUser(row)
	if err != nil {
		t.Fatalf("scanUser() error = %v", err)
	}
	if user.ID != "user-1" || user.City == nil || user.City.Name != "Москва" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if user.Role != "user" {
		t.Fatalf("unexpected role: %q", user.Role)
	}
	if user.AvatarURL == nil || *user.AvatarURL != "/uploads/avatars/file.png" {
		t.Fatalf("unexpected avatar: %+v", user.AvatarURL)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestScanUserParsesInterestTagIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, time.April, 15, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "email", "username", "password_hash", "birthday", "city_id", "avatar_url", "bio", "role", "interest_tag_ids", "created_at", "updated_at",
		"city_record_id", "city_name", "country_name", "timezone",
	}).AddRow(
		"user-1", "user@example.com", "user_1", "hash", nil, nil, nil, nil, "user",
		`["3dd89159-c28b-48bf-baa2-f920b0f315da","5f1f5ad0-343c-4e4a-afc4-3e2e68132f64"]`,
		now, now, nil, nil, nil, nil,
	)
	mock.ExpectQuery("SELECT \\* FROM user_account").WillReturnRows(rows)

	row := db.QueryRow("SELECT * FROM user_account")
	user, err := scanUser(row)
	if err != nil {
		t.Fatalf("scanUser() error = %v", err)
	}
	if len(user.InterestTagIDs) != 2 {
		t.Fatalf("interest tag ids count = %d, want 2: %#v", len(user.InterestTagIDs), user.InterestTagIDs)
	}
	if user.InterestTagIDs[0] != "3dd89159-c28b-48bf-baa2-f920b0f315da" {
		t.Fatalf("unexpected first interest tag id: %q", user.InterestTagIDs[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
