package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	authusecase "cityhawk/backend/internal/auth/usecase"
	"cityhawk/backend/internal/mocks"
	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
	"github.com/golang/mock/gomock"
)

func TestServiceIssueTokenPairSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mocks.NewMockUserReader(ctrl)
	refresh := mocks.NewMockRefreshSessionRepository(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)
	svc := authusecase.NewService(15*time.Minute, 24*time.Hour, users, refresh, tokens)

	user := usermodel.User{ID: "user-1", Email: "user@example.com"}
	tokens.EXPECT().SignAccessToken(user, 15*time.Minute).Return("access", nil)
	tokens.EXPECT().GenerateOpaqueToken(32).Return("refresh", nil)
	refresh.EXPECT().Store(gomock.Any(), "refresh", "user-1", gomock.Any()).Return(nil)

	pair, err := svc.IssueTokenPair(context.Background(), user)
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}
	if pair != (authmodel.TokenPair{AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", ExpiresIn: 900}) {
		t.Fatalf("unexpected token pair: %+v", pair)
	}
}

func TestServiceIssueTokenPairStoreFailureMapsToInternal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mocks.NewMockUserReader(ctrl)
	refresh := mocks.NewMockRefreshSessionRepository(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)
	svc := authusecase.NewService(time.Minute, time.Hour, users, refresh, tokens)
	user := usermodel.User{ID: "user-1", Email: "user@example.com"}

	tokens.EXPECT().SignAccessToken(user, time.Minute).Return("access", nil)
	tokens.EXPECT().GenerateOpaqueToken(32).Return("refresh", nil)
	refresh.EXPECT().Store(gomock.Any(), "refresh", "user-1", gomock.Any()).Return(errors.New("db failed"))

	_, err := svc.IssueTokenPair(context.Background(), user)
	if !errors.Is(err, platformerrors.ErrInternal) {
		t.Fatalf("error = %v, want ErrInternal", err)
	}
}

func TestServiceIssueTokenPairForUserIDNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mocks.NewMockUserReader(ctrl)
	refresh := mocks.NewMockRefreshSessionRepository(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)
	svc := authusecase.NewService(time.Minute, time.Hour, users, refresh, tokens)

	users.EXPECT().GetByID(gomock.Any(), "missing").Return(usermodel.User{}, false)

	_, err := svc.IssueTokenPairForUserID(context.Background(), "missing")
	if !errors.Is(err, platformerrors.ErrUserNotFound) {
		t.Fatalf("error = %v, want ErrUserNotFound", err)
	}
}

func TestServiceRotateRefreshAndRevoke(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mocks.NewMockUserReader(ctrl)
	refresh := mocks.NewMockRefreshSessionRepository(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)
	svc := authusecase.NewService(10*time.Minute, time.Hour, users, refresh, tokens)

	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(usermodel.User{ID: "user-1", Email: "u@example.com"}, true)
	refresh.EXPECT().Consume(gomock.Any(), "old-refresh").Return("user-1", nil)
	tokens.EXPECT().SignAccessToken(usermodel.User{ID: "user-1", Email: "u@example.com"}, 10*time.Minute).Return("new-access", nil)
	tokens.EXPECT().GenerateOpaqueToken(32).Return("new-refresh", nil)
	refresh.EXPECT().Store(gomock.Any(), "new-refresh", "user-1", gomock.Any()).Return(nil)
	refresh.EXPECT().Revoke(gomock.Any(), "revoke-me").Return(nil)

	pair, err := svc.RotateRefresh(context.Background(), "old-refresh")
	if err != nil {
		t.Fatalf("RotateRefresh() error = %v", err)
	}
	if pair.AccessToken != "new-access" || pair.RefreshToken != "new-refresh" {
		t.Fatalf("unexpected token pair: %+v", pair)
	}
	if err := svc.RevokeRefresh(context.Background(), "revoke-me"); err != nil {
		t.Fatalf("RevokeRefresh() error = %v", err)
	}
}

func TestServiceRotateRefreshConsumeFailureMapsToTokenRevoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mocks.NewMockUserReader(ctrl)
	refresh := mocks.NewMockRefreshSessionRepository(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)
	svc := authusecase.NewService(time.Minute, time.Hour, users, refresh, tokens)

	refresh.EXPECT().Consume(gomock.Any(), "bad").Return("", errors.New("nope"))

	_, err := svc.RotateRefresh(context.Background(), "bad")
	if !errors.Is(err, platformerrors.ErrTokenRevoked) {
		t.Fatalf("error = %v, want ErrTokenRevoked", err)
	}
}
