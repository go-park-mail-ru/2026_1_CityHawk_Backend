package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	platformerrors "cityhawk/backend/internal/platform/errors"
)

func TestInMemoryRefreshRepositoryStoreConsumeAndRevoke(t *testing.T) {
	t.Run("consume stored session", func(t *testing.T) {
		repo := NewInMemoryRefreshRepository()
		expiresAt := time.Now().UTC().Add(time.Minute)

		if err := repo.Store(context.Background(), "refresh-token", "user-1", expiresAt); err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		userID, err := repo.Consume(context.Background(), "refresh-token")
		if err != nil {
			t.Fatalf("Consume() error = %v", err)
		}
		if userID != "user-1" {
			t.Fatalf("Consume() userID = %q, want %q", userID, "user-1")
		}

		_, err = repo.Consume(context.Background(), "refresh-token")
		if !errors.Is(err, platformerrors.ErrTokenRevoked) {
			t.Fatalf("second Consume() error = %v, want ErrTokenRevoked", err)
		}
	})

	t.Run("consume expired session", func(t *testing.T) {
		repo := NewInMemoryRefreshRepository()

		if err := repo.Store(context.Background(), "expired-token", "user-2", time.Now().UTC().Add(-time.Minute)); err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		_, err := repo.Consume(context.Background(), "expired-token")
		if !errors.Is(err, platformerrors.ErrTokenExpired) {
			t.Fatalf("Consume() error = %v, want ErrTokenExpired", err)
		}

		_, err = repo.Consume(context.Background(), "expired-token")
		if !errors.Is(err, platformerrors.ErrTokenRevoked) {
			t.Fatalf("second Consume() error = %v, want ErrTokenRevoked", err)
		}
	})

	t.Run("revoke removes stored session", func(t *testing.T) {
		repo := NewInMemoryRefreshRepository()

		if err := repo.Store(context.Background(), "revoked-token", "user-3", time.Now().UTC().Add(time.Minute)); err != nil {
			t.Fatalf("Store() error = %v", err)
		}
		if err := repo.Revoke(context.Background(), "revoked-token"); err != nil {
			t.Fatalf("Revoke() error = %v", err)
		}

		_, err := repo.Consume(context.Background(), "revoked-token")
		if !errors.Is(err, platformerrors.ErrTokenRevoked) {
			t.Fatalf("Consume() error = %v, want ErrTokenRevoked", err)
		}
	})
}
