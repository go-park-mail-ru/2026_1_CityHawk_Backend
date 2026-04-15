package config

import "testing"

func TestDatabaseConfigDSNEscapesSpecialCharacters(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "db.example.com",
		Port:     "5432",
		User:     "user name",
		Password: "p@ss:/ word",
		Name:     "city hawk/db",
		SSLMode:  "verify-full",
	}

	got := cfg.DSN()
	want := "postgres://user%20name:p%40ss%3A%2F%20word@db.example.com:5432/city%20hawk%2Fdb?sslmode=verify-full"

	if got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}

func TestDatabaseConfigDSNFormatsIPv6Host(t *testing.T) {
	cfg := DatabaseConfig{
		Host:    "::1",
		Port:    "5432",
		User:    "postgres",
		Name:    "cityhawk",
		SSLMode: "disable",
	}

	got := cfg.DSN()
	want := "postgres://postgres@[::1]:5432/cityhawk?sslmode=disable"

	if got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}
