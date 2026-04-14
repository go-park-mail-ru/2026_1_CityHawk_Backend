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
		"id", "email", "username", "user_surname", "password_hash", "birthday", "city_id", "avatar_url", "created_at", "updated_at",
		"city_record_id", "city_name", "country_name", "timezone",
	}).AddRow(
		"user-1", "user@example.com", "user_1", "Иванов", "hash", now, "city-1", "/uploads/avatars/file.png", now, now,
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
	if user.AvatarURL == nil || *user.AvatarURL != "/uploads/avatars/file.png" {
		t.Fatalf("unexpected avatar: %+v", user.AvatarURL)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
