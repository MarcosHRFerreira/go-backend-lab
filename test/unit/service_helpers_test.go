package unit_test

import (
	"context"
	"testing"
	"time"

	"go-tweets/internal/apperror"
	"go-tweets/internal/config"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	jwtutil "go-tweets/pkg/internalsql/jwt"
)

type userRepositoryStub struct {
	getUserByEmailOrUsernameFunc func(ctx context.Context, email, username string) (*model.UserModel, error)
	createUserFunc               func(ctx context.Context, user *model.UserModel) (int, error)
	getRefreshTokenFunc          func(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error)
	storeRefreshTokenFunc        func(ctx context.Context, token *model.RefreshTokenModel) error
	getUserByIDFunc              func(ctx context.Context, userID int) (*model.UserModel, error)
	deleteRefreshTokenFunc       func(ctx context.Context, userID int) error
}

func (s *userRepositoryStub) GetUserByEmailOrUsername(ctx context.Context, email, username string) (*model.UserModel, error) {
	if s.getUserByEmailOrUsernameFunc == nil {
		return nil, nil
	}

	return s.getUserByEmailOrUsernameFunc(ctx, email, username)
}

func (s *userRepositoryStub) CreateUser(ctx context.Context, user *model.UserModel) (int, error) {
	if s.createUserFunc == nil {
		return 0, nil
	}

	return s.createUserFunc(ctx, user)
}

func (s *userRepositoryStub) GetRefreshToken(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error) {
	if s.getRefreshTokenFunc == nil {
		return nil, nil
	}

	return s.getRefreshTokenFunc(ctx, userID, now)
}

func (s *userRepositoryStub) StoreRefreshToken(ctx context.Context, token *model.RefreshTokenModel) error {
	if s.storeRefreshTokenFunc == nil {
		return nil
	}

	return s.storeRefreshTokenFunc(ctx, token)
}

func (s *userRepositoryStub) GetUserByID(ctx context.Context, userID int) (*model.UserModel, error) {
	if s.getUserByIDFunc == nil {
		return nil, nil
	}

	return s.getUserByIDFunc(ctx, userID)
}

func (s *userRepositoryStub) DeleteRefreshToken(ctx context.Context, userID int) error {
	if s.deleteRefreshTokenFunc == nil {
		return nil
	}

	return s.deleteRefreshTokenFunc(ctx, userID)
}

type postRepositoryStub struct {
	storePostFunc         func(ctx context.Context, post *model.PostModel) (int, error)
	getPostByIDFunc       func(ctx context.Context, postID int) (*model.PostWithUserModel, error)
	updatePostFunc        func(ctx context.Context, post *model.PostModel, postID int) error
	softDeletePostFunc    func(ctx context.Context, postID int, now time.Time) error
	isUserAlreadyLikeFunc func(ctx context.Context, postID, userID int) (bool, error)
	deleteLikePostFunc    func(ctx context.Context, postID, userID int) error
	storeLikePostFunc     func(ctx context.Context, like *model.PostLikeModel) error
	totalPostFunc         func(ctx context.Context) (int64, error)
	getAllPostFunc        func(ctx context.Context, param *dto.GetAllPostRequest, offset int) ([]model.PostWithUserModel, error)
}

func (s *postRepositoryStub) StorePost(ctx context.Context, post *model.PostModel) (int, error) {
	if s.storePostFunc == nil {
		return 0, nil
	}

	return s.storePostFunc(ctx, post)
}

func (s *postRepositoryStub) GetPostByID(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
	if s.getPostByIDFunc == nil {
		return nil, nil
	}

	return s.getPostByIDFunc(ctx, postID)
}

func (s *postRepositoryStub) UpdatePost(ctx context.Context, post *model.PostModel, postID int) error {
	if s.updatePostFunc == nil {
		return nil
	}

	return s.updatePostFunc(ctx, post, postID)
}

func (s *postRepositoryStub) SoftDeletePost(ctx context.Context, postID int, now time.Time) error {
	if s.softDeletePostFunc == nil {
		return nil
	}

	return s.softDeletePostFunc(ctx, postID, now)
}

func (s *postRepositoryStub) IsUserAlreadyLikePost(ctx context.Context, postID, userID int) (bool, error) {
	if s.isUserAlreadyLikeFunc == nil {
		return false, nil
	}

	return s.isUserAlreadyLikeFunc(ctx, postID, userID)
}

func (s *postRepositoryStub) DeleteLikePost(ctx context.Context, postID, userID int) error {
	if s.deleteLikePostFunc == nil {
		return nil
	}

	return s.deleteLikePostFunc(ctx, postID, userID)
}

func (s *postRepositoryStub) StoreLikePost(ctx context.Context, like *model.PostLikeModel) error {
	if s.storeLikePostFunc == nil {
		return nil
	}

	return s.storeLikePostFunc(ctx, like)
}

func (s *postRepositoryStub) TotalPost(ctx context.Context) (int64, error) {
	if s.totalPostFunc == nil {
		return 0, nil
	}

	return s.totalPostFunc(ctx)
}

func (s *postRepositoryStub) GetAllPost(ctx context.Context, param *dto.GetAllPostRequest, offset int) ([]model.PostWithUserModel, error) {
	if s.getAllPostFunc == nil {
		return nil, nil
	}

	return s.getAllPostFunc(ctx, param, offset)
}

type commentRepositoryStub struct {
	storeCommentFunc         func(ctx context.Context, comment *model.CommentModel) error
	detailCommentFunc        func(ctx context.Context, commentID int) (*model.CommentModel, error)
	isUserAlreadyLikeFunc    func(ctx context.Context, commentID, userID int) (bool, error)
	deleteLikeCommentFunc    func(ctx context.Context, commentID, userID int) error
	storeLikeCommentFunc     func(ctx context.Context, like *model.CommentLikeModel) error
	getCommentsByPostIDsFunc func(ctx context.Context, postIDs []int) ([]model.CommentModel, error)
}

func (s *commentRepositoryStub) StoreComment(ctx context.Context, comment *model.CommentModel) error {
	if s.storeCommentFunc == nil {
		return nil
	}

	return s.storeCommentFunc(ctx, comment)
}

func (s *commentRepositoryStub) DetailComment(ctx context.Context, commentID int) (*model.CommentModel, error) {
	if s.detailCommentFunc == nil {
		return nil, nil
	}

	return s.detailCommentFunc(ctx, commentID)
}

func (s *commentRepositoryStub) IsUserAlreadyLikeComment(ctx context.Context, commentID, userID int) (bool, error) {
	if s.isUserAlreadyLikeFunc == nil {
		return false, nil
	}

	return s.isUserAlreadyLikeFunc(ctx, commentID, userID)
}

func (s *commentRepositoryStub) DeleteLikeComment(ctx context.Context, commentID, userID int) error {
	if s.deleteLikeCommentFunc == nil {
		return nil
	}

	return s.deleteLikeCommentFunc(ctx, commentID, userID)
}

func (s *commentRepositoryStub) StoreLikeComment(ctx context.Context, like *model.CommentLikeModel) error {
	if s.storeLikeCommentFunc == nil {
		return nil
	}

	return s.storeLikeCommentFunc(ctx, like)
}

func (s *commentRepositoryStub) GetCommentsByPostIDs(ctx context.Context, postIDs []int) ([]model.CommentModel, error) {
	if s.getCommentsByPostIDsFunc == nil {
		return nil, nil
	}

	return s.getCommentsByPostIDsFunc(ctx, postIDs)
}

func newTestConfig() *config.Config {
	return &config.Config{
		SecretJwt: "test-secret",
	}
}

func assertStatusCode(t *testing.T, err error, expected int) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error with status %d, got nil", expected)
	}

	if actual := apperror.StatusCode(err); actual != expected {
		t.Fatalf("expected status %d, got %d", expected, actual)
	}
}

func assertTokenClaims(t *testing.T, token string, expectedUserID int, expectedUsername string) {
	t.Helper()

	userID, username, err := jwtutil.ValidateToken(token, newTestConfig().SecretJwt, true)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if userID != expectedUserID {
		t.Fatalf("expected user id %d, got %d", expectedUserID, userID)
	}

	if username != expectedUsername {
		t.Fatalf("expected username %q, got %q", expectedUsername, username)
	}
}
