package usecase_test

import (
	"context"
	"errors"
	"testing"

	authmodel "cityhawk/backend/internal/auth/model"
	authusecase "cityhawk/backend/internal/auth/usecase"
	"cityhawk/backend/internal/mocks"
	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
	"github.com/golang/mock/gomock"
)

func TestOAuthUserServiceFindOrCreateFromOAuth(t *testing.T) {
	t.Run("returns existing user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		users := mocks.NewMockAuthUserRepository(ctrl)
		passwords := mocks.NewMockPasswordService(ctrl)
		svc := authusecase.NewOAuthUserService(users, passwords)
		existing := usermodel.User{ID: "user-1", Email: "user@example.com"}

		users.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(existing, true)

		user, err := svc.FindOrCreateFromOAuth(context.Background(), authmodel.OAuthIdentity{
			Provider:  "yandex",
			SubjectID: "sub-1",
			Email:     "User@example.com",
			Username:  "User Name",
		})
		if err != nil {
			t.Fatalf("FindOrCreateFromOAuth() error = %v", err)
		}
		if user.ID != "user-1" {
			t.Fatalf("unexpected user: %+v", user)
		}
	})

	t.Run("creates new user with generated local email", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		users := mocks.NewMockAuthUserRepository(ctrl)
		passwords := mocks.NewMockPasswordService(ctrl)
		svc := authusecase.NewOAuthUserService(users, passwords)

		users.EXPECT().GetByEmail(gomock.Any(), "yandex_sub1@yandex.local").Return(usermodel.User{}, false)
		passwords.EXPECT().Hash(gomock.Any()).Return("hashed", nil)
		users.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(usermodel.User{})).
			DoAndReturn(func(_ context.Context, u usermodel.User) (usermodel.User, error) {
				if u.Email != "yandex_sub1@yandex.local" || u.Username != "sub1" || u.UserSurname != "sub1" {
					t.Fatalf("unexpected created user: %+v", u)
				}
				u.ID = "user-2"
				return u, nil
			})

		user, err := svc.FindOrCreateFromOAuth(context.Background(), authmodel.OAuthIdentity{
			Provider:  "yandex",
			SubjectID: "sub-1",
			Username:  "sub-1",
		})
		if err != nil {
			t.Fatalf("FindOrCreateFromOAuth() error = %v", err)
		}
		if user.ID != "user-2" {
			t.Fatalf("unexpected user: %+v", user)
		}
	})

	t.Run("maps email race to existing user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		users := mocks.NewMockAuthUserRepository(ctrl)
		passwords := mocks.NewMockPasswordService(ctrl)
		svc := authusecase.NewOAuthUserService(users, passwords)
		existing := usermodel.User{ID: "user-3", Email: "user@example.com"}

		users.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(usermodel.User{}, false)
		passwords.EXPECT().Hash(gomock.Any()).Return("hashed", nil)
		users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(usermodel.User{}, platformerrors.ErrEmailExists)
		users.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(existing, true)

		user, err := svc.FindOrCreateFromOAuth(context.Background(), authmodel.OAuthIdentity{
			Provider:  "yandex",
			SubjectID: "sub-1",
			Email:     "user@example.com",
			Username:  "user",
		})
		if err != nil {
			t.Fatalf("FindOrCreateFromOAuth() error = %v", err)
		}
		if user.ID != "user-3" {
			t.Fatalf("unexpected user: %+v", user)
		}
	})
}

func TestOAuthUserServiceValidationAndHashErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mocks.NewMockAuthUserRepository(ctrl)
	passwords := mocks.NewMockPasswordService(ctrl)
	svc := authusecase.NewOAuthUserService(users, passwords)

	if _, err := svc.FindOrCreateFromOAuth(context.Background(), authmodel.OAuthIdentity{}); err == nil {
		t.Fatal("expected provider error")
	}
	if _, err := svc.FindOrCreateFromOAuth(context.Background(), authmodel.OAuthIdentity{Provider: "yandex"}); err == nil {
		t.Fatal("expected empty identity error")
	}

	users.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(usermodel.User{}, false)
	passwords.EXPECT().Hash(gomock.Any()).Return("", errors.New("hash failed"))
	if _, err := svc.FindOrCreateFromOAuth(context.Background(), authmodel.OAuthIdentity{
		Provider:  "yandex",
		SubjectID: "sub-1",
		Email:     "user@example.com",
	}); err == nil {
		t.Fatal("expected hash error")
	}
}
