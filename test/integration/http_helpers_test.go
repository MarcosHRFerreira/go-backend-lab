package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"go-tweets/internal/config"
	"go-tweets/internal/dto"
	commenthandler "go-tweets/internal/handler/comment"
	posthandler "go-tweets/internal/handler/post"
	userhandler "go-tweets/internal/handler/user"
	commentservice "go-tweets/internal/service/comment"
	postservice "go-tweets/internal/service/post"
	userservice "go-tweets/internal/service/user"
	jwtutil "go-tweets/pkg/internalsql/jwt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type userServiceStub struct {
	registerFunc     func(ctx context.Context, req *dto.RegisterRequest) (int, error)
	loginFunc        func(ctx context.Context, req *dto.LoginRequest) (string, string, error)
	refreshTokenFunc func(ctx context.Context, req *dto.RefreshTokenRequest, userID int) (string, string, error)
}

func (s *userServiceStub) Register(ctx context.Context, req *dto.RegisterRequest) (int, error) {
	if s.registerFunc == nil {
		return 0, nil
	}

	return s.registerFunc(ctx, req)
}

func (s *userServiceStub) Login(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
	if s.loginFunc == nil {
		return "", "", nil
	}

	return s.loginFunc(ctx, req)
}

func (s *userServiceStub) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, userID int) (string, string, error) {
	if s.refreshTokenFunc == nil {
		return "", "", nil
	}

	return s.refreshTokenFunc(ctx, req, userID)
}

type postServiceStub struct {
	createPostFunc func(ctx context.Context, req *dto.CreateOrUpdatePostRequest, userID int) (int, error)
	updatePostFunc func(ctx context.Context, req *dto.CreateOrUpdatePostRequest, postID, userID int) error
	deletePostFunc func(ctx context.Context, postID, userID int) error
	likePostFunc   func(ctx context.Context, postID, userID int) error
	detailPostFunc func(ctx context.Context, postID int) (*dto.DetailPostResponse, error)
	getAllPostFunc func(ctx context.Context, param *dto.GetAllPostRequest) (*dto.GetAllPostResponse, error)
}

func (s *postServiceStub) CreatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, userID int) (int, error) {
	if s.createPostFunc == nil {
		return 0, nil
	}

	return s.createPostFunc(ctx, req, userID)
}

func (s *postServiceStub) UpdatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, postID, userID int) error {
	if s.updatePostFunc == nil {
		return nil
	}

	return s.updatePostFunc(ctx, req, postID, userID)
}

func (s *postServiceStub) DeletePost(ctx context.Context, postID, userID int) error {
	if s.deletePostFunc == nil {
		return nil
	}

	return s.deletePostFunc(ctx, postID, userID)
}

func (s *postServiceStub) LikeOrUnlikePost(ctx context.Context, postID, userID int) error {
	if s.likePostFunc == nil {
		return nil
	}

	return s.likePostFunc(ctx, postID, userID)
}

func (s *postServiceStub) DetailPost(ctx context.Context, postID int) (*dto.DetailPostResponse, error) {
	if s.detailPostFunc == nil {
		return nil, nil
	}

	return s.detailPostFunc(ctx, postID)
}

func (s *postServiceStub) GetAllPost(ctx context.Context, param *dto.GetAllPostRequest) (*dto.GetAllPostResponse, error) {
	if s.getAllPostFunc == nil {
		return nil, nil
	}

	return s.getAllPostFunc(ctx, param)
}

type commentServiceStub struct {
	createCommentFunc       func(ctx context.Context, req *dto.StoreCommentRequest, userID int) error
	likeOrUnlikeCommentFunc func(ctx context.Context, commentID, userID int) error
}

func (s *commentServiceStub) CreateComment(ctx context.Context, req *dto.StoreCommentRequest, userID int) error {
	if s.createCommentFunc == nil {
		return nil
	}

	return s.createCommentFunc(ctx, req, userID)
}

func (s *commentServiceStub) LikeOrUnLikeComment(ctx context.Context, commentID, userID int) error {
	if s.likeOrUnlikeCommentFunc == nil {
		return nil
	}

	return s.likeOrUnlikeCommentFunc(ctx, commentID, userID)
}

func newIntegrationConfig() *config.Config {
	return &config.Config{
		SecretJwt: "integration-secret",
	}
}

func newTestRouter(userSvc userservice.UserService, postSvc postservice.PostService, commentSvc commentservice.CommentService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	validate := validator.New()
	cfg := newIntegrationConfig()

	userhandler.NewHandler(router, validate, userSvc).RouteList(cfg.SecretJwt)
	posthandler.NewHandler(router, validate, postSvc).RouteList(cfg.SecretJwt)
	commenthandler.NewHandler(router, validate, commentSvc).RouteList(cfg.SecretJwt)

	return router
}

func performJSONRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody []byte
	var err error
	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder
}

func decodeJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var payload map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	return payload
}

func mustCreateToken(t *testing.T, userID int, username string) string {
	t.Helper()

	token, err := jwtutil.CreateToken(userID, username, newIntegrationConfig().SecretJwt)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	return token
}
