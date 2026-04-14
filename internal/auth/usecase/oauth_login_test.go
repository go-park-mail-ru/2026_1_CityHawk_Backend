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

func TestOAuthLoginServiceLoginWithGoogleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	google := mocks.NewMockOAuthGateway(ctrl)
	yandex := mocks.NewMockOAuthGateway(ctrl)
	vk := mocks.NewMockOAuthGateway(ctrl)
	users := mocks.NewMockOAuthUserFinder(ctrl)
	issuer := mocks.NewMockOAuthTokenIssuer(ctrl)
	svc := authusecase.NewOAuthLoginService(google, yandex, vk, users, issuer)

	identity := authmodel.OAuthIdentity{Email: "user@example.com", SubjectID: "sub-1"}
	user := usermodel.User{ID: "user-1", Email: "user@example.com"}

	google.EXPECT().FetchIdentity(gomock.Any(), "code").Return(identity, nil)
	users.EXPECT().FindOrCreateFromOAuth(gomock.Any(), identity).Return(user, nil)
	issuer.EXPECT().IssueTokenPair(gomock.Any(), user).Return(authmodel.TokenPair{AccessToken: "token"}, nil)

	pair, err := svc.LoginWithGoogle(context.Background(), "code")
	if err != nil {
		t.Fatalf("LoginWithGoogle() error = %v", err)
	}
	if pair.AccessToken != "token" {
		t.Fatalf("unexpected token pair: %+v", pair)
	}
}

func TestOAuthLoginServiceMapsErrors(t *testing.T) {
	t.Run("gateway error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		google := mocks.NewMockOAuthGateway(ctrl)
		yandex := mocks.NewMockOAuthGateway(ctrl)
		vk := mocks.NewMockOAuthGateway(ctrl)
		users := mocks.NewMockOAuthUserFinder(ctrl)
		issuer := mocks.NewMockOAuthTokenIssuer(ctrl)
		svc := authusecase.NewOAuthLoginService(google, yandex, vk, users, issuer)

		google.EXPECT().FetchIdentity(gomock.Any(), "bad-code").Return(authmodel.OAuthIdentity{}, errors.New("oauth failed"))
		_, err := svc.LoginWithGoogle(context.Background(), "bad-code")
		if err == nil {
			t.Fatal("expected oauth error")
		}
	})

	t.Run("user persistence failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		google := mocks.NewMockOAuthGateway(ctrl)
		yandex := mocks.NewMockOAuthGateway(ctrl)
		vk := mocks.NewMockOAuthGateway(ctrl)
		users := mocks.NewMockOAuthUserFinder(ctrl)
		issuer := mocks.NewMockOAuthTokenIssuer(ctrl)
		svc := authusecase.NewOAuthLoginService(google, yandex, vk, users, issuer)

		identity := authmodel.OAuthIdentity{Email: "user@example.com", SubjectID: "sub-1"}
		yandex.EXPECT().FetchIdentity(gomock.Any(), "code").Return(identity, nil)
		users.EXPECT().FindOrCreateFromOAuth(gomock.Any(), identity).Return(usermodel.User{}, errors.New("db failed"))
		_, err := svc.LoginWithYandex(context.Background(), "code")
		if !errors.Is(err, platformerrors.ErrInternal) {
			t.Fatalf("error = %v, want ErrInternal", err)
		}
	})

	t.Run("issuer failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		google := mocks.NewMockOAuthGateway(ctrl)
		yandex := mocks.NewMockOAuthGateway(ctrl)
		vk := mocks.NewMockOAuthGateway(ctrl)
		users := mocks.NewMockOAuthUserFinder(ctrl)
		issuer := mocks.NewMockOAuthTokenIssuer(ctrl)
		svc := authusecase.NewOAuthLoginService(google, yandex, vk, users, issuer)

		identity := authmodel.OAuthIdentity{Email: "user@example.com", SubjectID: "sub-1"}
		user := usermodel.User{ID: "user-1", Email: "user@example.com"}
		vk.EXPECT().FetchIdentity(gomock.Any(), "vk-code").Return(identity, nil)
		users.EXPECT().FindOrCreateFromOAuth(gomock.Any(), identity).Return(user, nil)
		issuer.EXPECT().IssueTokenPair(gomock.Any(), user).Return(authmodel.TokenPair{}, errors.New("issue failed"))
		_, err := svc.LoginWithVK(context.Background(), "vk-code")
		if !errors.Is(err, platformerrors.ErrIssueTokens) {
			t.Fatalf("error = %v, want ErrIssueTokens", err)
		}
	})
}
