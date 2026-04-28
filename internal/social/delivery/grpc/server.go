package grpc

import (
	"context"

	"cityhawk/backend/internal/grpcconv"
	platformerrors "cityhawk/backend/internal/platform/errors"
	socialmodel "cityhawk/backend/internal/social/model"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	socialv1 "cityhawk/backend/pkg/pb/social/v1"
)

const defaultPageLimit = 20

type Repository interface {
	AddFavorite(ctx context.Context, userID, eventID string) error
	RemoveFavorite(ctx context.Context, userID, eventID string) error
	ListFavoriteEvents(ctx context.Context, userID string, limit, offset int) ([]socialmodel.FavoriteEvent, error)
	IsFavorite(ctx context.Context, userID, eventID string) (socialmodel.FavoriteFlag, error)
	FollowUser(ctx context.Context, followerUserID, followedUserID string) error
	UnfollowUser(ctx context.Context, followerUserID, followedUserID string) error
	ListFollowers(ctx context.Context, userID string, limit, offset int) ([]socialmodel.UserFollow, error)
	ListFollowing(ctx context.Context, userID string, limit, offset int) ([]socialmodel.UserFollow, error)
	IsFollowing(ctx context.Context, followerUserID, followedUserID string) (socialmodel.FollowingFlag, error)
	FavoriteFlags(ctx context.Context, userID string, eventIDs []string) (map[string]socialmodel.FavoriteFlag, error)
	FollowingFlags(ctx context.Context, followerUserID string, followedUserIDs []string) (map[string]socialmodel.FollowingFlag, error)
}

type Server struct {
	socialv1.UnimplementedSocialServiceServer

	repo Repository
}

func NewServer(repo Repository) *Server {
	return &Server{repo: repo}
}

func (s *Server) AddFavorite(ctx context.Context, req *socialv1.AddFavoriteRequest) (*commonv1.BoolResponse, error) {
	userID, err := requiredUserID(req.GetActor())
	if err != nil {
		return nil, err
	}
	if err := s.repo.AddFavorite(ctx, userID, req.GetEventId()); err != nil {
		return nil, grpcconv.Error(err)
	}
	return &commonv1.BoolResponse{Ok: true}, nil
}

func (s *Server) RemoveFavorite(ctx context.Context, req *socialv1.RemoveFavoriteRequest) (*commonv1.BoolResponse, error) {
	userID, err := requiredUserID(req.GetActor())
	if err != nil {
		return nil, err
	}
	if err := s.repo.RemoveFavorite(ctx, userID, req.GetEventId()); err != nil {
		return nil, grpcconv.Error(err)
	}
	return &commonv1.BoolResponse{Ok: true}, nil
}

func (s *Server) ListFavoriteEvents(ctx context.Context, req *socialv1.ListFavoriteEventsRequest) (*socialv1.ListFavoriteEventsResponse, error) {
	userID, err := requiredUserID(req.GetActor())
	if err != nil {
		return nil, err
	}
	limit, offset := page(req.GetPage())
	items, err := s.repo.ListFavoriteEvents(ctx, userID, limit, offset)
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	response := &socialv1.ListFavoriteEventsResponse{
		Items: make([]*socialv1.FavoriteEvent, 0, len(items)),
		Page:  pageResponse(len(items), limit, offset),
	}
	for _, item := range items {
		response.Items = append(response.Items, favoriteEventToProto(item))
	}
	return response, nil
}

func (s *Server) IsFavorite(ctx context.Context, req *socialv1.IsFavoriteRequest) (*socialv1.IsFavoriteResponse, error) {
	flag, err := s.repo.IsFavorite(ctx, req.GetUserId(), req.GetEventId())
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	return &socialv1.IsFavoriteResponse{IsFavorite: flag.IsFavorite}, nil
}

func (s *Server) FollowUser(ctx context.Context, req *socialv1.FollowUserRequest) (*commonv1.BoolResponse, error) {
	userID, err := requiredUserID(req.GetActor())
	if err != nil {
		return nil, err
	}
	if err := s.repo.FollowUser(ctx, userID, req.GetFollowedUserId()); err != nil {
		return nil, grpcconv.Error(err)
	}
	return &commonv1.BoolResponse{Ok: true}, nil
}

func (s *Server) UnfollowUser(ctx context.Context, req *socialv1.UnfollowUserRequest) (*commonv1.BoolResponse, error) {
	userID, err := requiredUserID(req.GetActor())
	if err != nil {
		return nil, err
	}
	if err := s.repo.UnfollowUser(ctx, userID, req.GetFollowedUserId()); err != nil {
		return nil, grpcconv.Error(err)
	}
	return &commonv1.BoolResponse{Ok: true}, nil
}

func (s *Server) ListFollowers(ctx context.Context, req *socialv1.ListFollowersRequest) (*socialv1.ListUsersResponse, error) {
	limit, offset := page(req.GetPage())
	items, err := s.repo.ListFollowers(ctx, req.GetUserId(), limit, offset)
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	return usersResponse(ctx, s.repo, grpcconv.UserIDFromContext(req.GetViewerContext()), items, limit, offset)
}

func (s *Server) ListFollowing(ctx context.Context, req *socialv1.ListFollowingRequest) (*socialv1.ListUsersResponse, error) {
	limit, offset := page(req.GetPage())
	items, err := s.repo.ListFollowing(ctx, req.GetUserId(), limit, offset)
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	return usersResponse(ctx, s.repo, grpcconv.UserIDFromContext(req.GetViewerContext()), items, limit, offset)
}

func (s *Server) IsFollowing(ctx context.Context, req *socialv1.IsFollowingRequest) (*socialv1.IsFollowingResponse, error) {
	flag, err := s.repo.IsFollowing(ctx, req.GetFollowerUserId(), req.GetFollowedUserId())
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	return &socialv1.IsFollowingResponse{IsFollowing: flag.IsFollowing}, nil
}

func (s *Server) GetSocialFlags(ctx context.Context, req *socialv1.GetSocialFlagsRequest) (*socialv1.SocialFlagsResponse, error) {
	viewerID := grpcconv.UserIDFromContext(req.GetViewerContext())
	favorites, err := s.repo.FavoriteFlags(ctx, viewerID, req.GetEventIds())
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	following, err := s.repo.FollowingFlags(ctx, viewerID, req.GetAuthorUserIds())
	if err != nil {
		return nil, grpcconv.Error(err)
	}

	response := &socialv1.SocialFlagsResponse{
		FavoriteEvents: make([]*socialv1.EventFavoriteFlag, 0, len(favorites)),
		FollowingUsers: make([]*socialv1.UserFollowingFlag, 0, len(following)),
	}
	for _, eventID := range req.GetEventIds() {
		response.FavoriteEvents = append(response.FavoriteEvents, favoriteFlagToProto(favorites[eventID]))
	}
	for _, userID := range req.GetAuthorUserIds() {
		response.FollowingUsers = append(response.FollowingUsers, followingFlagToProto(following[userID]))
	}
	return response, nil
}

func usersResponse(ctx context.Context, repo Repository, viewerID string, items []socialmodel.UserFollow, limit, offset int) (*socialv1.ListUsersResponse, error) {
	userIDs := make([]string, 0, len(items))
	for _, item := range items {
		userIDs = append(userIDs, item.UserID)
	}
	flags, err := repo.FollowingFlags(ctx, viewerID, userIDs)
	if err != nil {
		return nil, grpcconv.Error(err)
	}

	response := &socialv1.ListUsersResponse{
		Items: make([]*socialv1.UserFollow, 0, len(items)),
		Page:  pageResponse(len(items), limit, offset),
	}
	for _, item := range items {
		response.Items = append(response.Items, &socialv1.UserFollow{
			UserId:      item.UserID,
			IsFollowing: flags[item.UserID].IsFollowing,
			CreatedAt:   grpcconv.TimeToProto(item.CreatedAt),
		})
	}
	return response, nil
}

func favoriteEventToProto(item socialmodel.FavoriteEvent) *socialv1.FavoriteEvent {
	return &socialv1.FavoriteEvent{
		EventId:   item.EventID,
		CreatedAt: grpcconv.TimeToProto(item.CreatedAt),
	}
}

func favoriteFlagToProto(flag socialmodel.FavoriteFlag) *socialv1.EventFavoriteFlag {
	return &socialv1.EventFavoriteFlag{
		EventId:    flag.EventID,
		IsFavorite: flag.IsFavorite,
		CreatedAt:  grpcconv.OptionalTimeToProto(flag.CreatedAt),
	}
}

func followingFlagToProto(flag socialmodel.FollowingFlag) *socialv1.UserFollowingFlag {
	return &socialv1.UserFollowingFlag{
		UserId:      flag.UserID,
		IsFollowing: flag.IsFollowing,
		CreatedAt:   grpcconv.OptionalTimeToProto(flag.CreatedAt),
	}
}

func requiredUserID(ctx *commonv1.UserContext) (string, error) {
	userID := grpcconv.UserIDFromContext(ctx)
	if userID == "" {
		return "", grpcconv.Error(platformerrors.ErrUnauthorized)
	}
	return userID, nil
}

func page(req *commonv1.PageRequest) (int, int) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = defaultPageLimit
	}
	offset := int(req.GetOffset())
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func pageResponse(total, limit, offset int) *commonv1.PageResponse {
	return &commonv1.PageResponse{
		Total:  int32(total),
		Limit:  int32(limit),
		Offset: int32(offset),
	}
}
