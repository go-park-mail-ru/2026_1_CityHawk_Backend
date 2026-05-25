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

func TestAuthFlowServiceRegisterSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mocks.NewMockAuthUserRepository(ctrl)
	issuer := mocks.NewMocktokenPairIssuer(ctrl)
	passwords := mocks.NewMockPasswordService(ctrl)
	svc := authusecase.NewAuthFlowService(users, issuer, passwords)

	passwords.EXPECT().Hash("verysecret").Return("hashed", nil)
	users.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(usermodel.User{})).
		DoAndReturn(func(_ context.Context, u usermodel.User) (usermodel.User, error) {
			if u.Email != "tester@example.com" || u.PasswordHash != "hashed" {
				t.Fatalf("unexpected persisted user: %+v", u)
			}
			if u.Username != "" || u.UserSurname != "" || u.Birthday != nil || u.CityID != nil {
				t.Fatalf("profile fields should be empty on register: %+v", u)
			}
			u.ID = "user-1"
			return u, nil
		})
	issuer.EXPECT().IssueTokenPair(gomock.Any(), gomock.AssignableToTypeOf(usermodel.User{})).
		Return(authmodel.TokenPair{AccessToken: "access", RefreshToken: "refresh"}, nil)

	res, err := svc.Register(context.Background(), authmodel.RegisterInput{
		Email:    " Tester@example.com ",
		Password: "verysecret",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if res.User.ID != "user-1" || res.Tokens.AccessToken != "access" {
		t.Fatalf("unexpected register result: %+v", res)
	}
}

func TestAuthFlowServiceRegisterMapsErrors(t *testing.T) {
	t.Run("hash failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		users := mocks.NewMockAuthUserRepository(ctrl)
		issuer := mocks.NewMocktokenPairIssuer(ctrl)
		passwords := mocks.NewMockPasswordService(ctrl)
		svc := authusecase.NewAuthFlowService(users, issuer, passwords)

		passwords.EXPECT().Hash("verysecret").Return("", errors.New("hash failed"))
		_, err := svc.Register(context.Background(), authmodel.RegisterInput{
			Email:    "user@example.com",
			Password: "verysecret",
		})
		if !errors.Is(err, platformerrors.ErrInternal) {
			t.Fatalf("error = %v, want ErrInternal", err)
		}
	})

	t.Run("email exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		users := mocks.NewMockAuthUserRepository(ctrl)
		issuer := mocks.NewMocktokenPairIssuer(ctrl)
		passwords := mocks.NewMockPasswordService(ctrl)
		svc := authusecase.NewAuthFlowService(users, issuer, passwords)

		passwords.EXPECT().Hash("verysecret").Return("hashed", nil)
		users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(usermodel.User{}, platformerrors.ErrEmailExists)
		_, err := svc.Register(context.Background(), authmodel.RegisterInput{
			Email:    "user@example.com",
			Password: "verysecret",
		})
		if !errors.Is(err, platformerrors.ErrEmailExists) {
			t.Fatalf("error = %v, want ErrEmailExists", err)
		}
	})
}

func TestAuthFlowServiceLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mocks.NewMockAuthUserRepository(ctrl)
	issuer := mocks.NewMocktokenPairIssuer(ctrl)
	passwords := mocks.NewMockPasswordService(ctrl)
	svc := authusecase.NewAuthFlowService(users, issuer, passwords)
	user := usermodel.User{ID: "user-1", Email: "user@example.com", PasswordHash: "stored"}

	users.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(user, true)
	passwords.EXPECT().Verify("verysecret", "stored").Return(true)
	issuer.EXPECT().IssueTokenPair(gomock.Any(), user).Return(authmodel.TokenPair{AccessToken: "access"}, nil)

	res, err := svc.Login(context.Background(), authmodel.LoginInput{Email: "user@example.com", Password: "verysecret"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if res.User.ID != "user-1" || res.Tokens.AccessToken != "access" {
		t.Fatalf("unexpected login result: %+v", res)
	}
}

func TestAuthFlowServiceLoginInvalidCredentialsAndIssueError(t *testing.T) {
	t.Run("invalid credentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		users := mocks.NewMockAuthUserRepository(ctrl)
		issuer := mocks.NewMocktokenPairIssuer(ctrl)
		passwords := mocks.NewMockPasswordService(ctrl)
		svc := authusecase.NewAuthFlowService(users, issuer, passwords)

		users.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(usermodel.User{}, false)
		_, err := svc.Login(context.Background(), authmodel.LoginInput{Email: "user@example.com", Password: "verysecret"})
		if !errors.Is(err, platformerrors.ErrInvalidCredentials) {
			t.Fatalf("error = %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("issuer failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		users := mocks.NewMockAuthUserRepository(ctrl)
		issuer := mocks.NewMocktokenPairIssuer(ctrl)
		passwords := mocks.NewMockPasswordService(ctrl)
		svc := authusecase.NewAuthFlowService(users, issuer, passwords)
		user := usermodel.User{ID: "user-1", Email: "user@example.com", PasswordHash: "stored"}

		users.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(user, true)
		passwords.EXPECT().Verify("verysecret", "stored").Return(true)
		issuer.EXPECT().IssueTokenPair(gomock.Any(), user).Return(authmodel.TokenPair{}, errors.New("boom"))
		_, err := svc.Login(context.Background(), authmodel.LoginInput{Email: "user@example.com", Password: "verysecret"})
		if !errors.Is(err, platformerrors.ErrIssueTokens) {
			t.Fatalf("error = %v, want ErrIssueTokens", err)
		}
	})
}
