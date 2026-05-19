package unit_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	postservice "go-tweets/internal/service/post"
)

func TestPostServiceCreatePostSuccess(t *testing.T) {
	t.Parallel()

	var storedPost *model.PostModel
	service := postservice.NewPostService(newTestConfig(), &postRepositoryStub{
		storePostFunc: func(ctx context.Context, post *model.PostModel) (int, error) {
			storedPost = post
			return 15, nil
		},
	}, &commentRepositoryStub{})

	postID, err := service.CreatePost(context.Background(), &dto.CreateOrUpdatePostRequest{
		Title:   "hello",
		Content: "market level go",
	}, 3)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if postID != 15 {
		t.Fatalf("expected post id 15, got %d", postID)
	}

	if storedPost == nil {
		t.Fatal("expected post to be persisted")
	}

	if storedPost.UserID != 3 {
		t.Fatalf("expected user id 3, got %d", storedPost.UserID)
	}

	if storedPost.Title != "hello" || storedPost.Content != "market level go" {
		t.Fatalf("unexpected stored post payload: %+v", storedPost)
	}
}

func TestPostServiceGetAllPostAggregatesCommentsAndPagination(t *testing.T) {
	t.Parallel()

	service := postservice.NewPostService(newTestConfig(), &postRepositoryStub{
		totalPostFunc: func(ctx context.Context) (int64, error) {
			return 3, nil
		},
		getAllPostFunc: func(ctx context.Context, param *dto.GetAllPostRequest, offset int) ([]model.PostWithUserModel, error) {
			if offset != 0 {
				t.Fatalf("expected offset 0, got %d", offset)
			}

			return []model.PostWithUserModel{
				{
					ID:        1,
					UserID:    8,
					Username:  "marcos",
					Title:     "first",
					Content:   "content",
					LikeCount: 2,
					CreatedAt: time.Unix(100, 0).UTC(),
					UpdatedAt: time.Unix(200, 0).UTC(),
				},
				{
					ID:        2,
					UserID:    9,
					Username:  "ana",
					Title:     "second",
					Content:   "content 2",
					LikeCount: 1,
					CreatedAt: time.Unix(300, 0).UTC(),
					UpdatedAt: time.Unix(400, 0).UTC(),
				},
			}, nil
		},
	}, &commentRepositoryStub{
		getCommentsByPostIDsFunc: func(ctx context.Context, postIDs []int) ([]model.CommentModel, error) {
			if len(postIDs) != 2 {
				t.Fatalf("expected 2 post ids, got %d", len(postIDs))
			}

			return []model.CommentModel{
				{
					ID:        10,
					PostID:    1,
					Username:  "joao",
					Content:   "nice",
					LikeCount: 5,
					CreatedAt: time.Unix(500, 0).UTC(),
					UpdatedAt: time.Unix(600, 0).UTC(),
				},
			}, nil
		},
	})

	result, err := service.GetAllPost(context.Background(), &dto.GetAllPostRequest{
		Page:  1,
		Limit: 2,
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if result.TotalPage != 2 {
		t.Fatalf("expected total page 2, got %d", result.TotalPage)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result.Data))
	}

	if len(result.Data[0].Comments) != 1 {
		t.Fatalf("expected 1 comment on first post, got %d", len(result.Data[0].Comments))
	}

	if len(result.Data[1].Comments) != 0 {
		t.Fatalf("expected 0 comments on second post, got %d", len(result.Data[1].Comments))
	}
}

func TestPostServiceDetailPostReturnsNotFoundWhenPostDoesNotExist(t *testing.T) {
	t.Parallel()

	service := postservice.NewPostService(newTestConfig(), &postRepositoryStub{
		getPostByIDFunc: func(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
			return nil, nil
		},
	}, &commentRepositoryStub{})

	_, err := service.DetailPost(context.Background(), 20)

	assertStatusCode(t, err, http.StatusNotFound)

	if err.Error() != "post not found" {
		t.Fatalf("expected post not found error, got %q", err.Error())
	}
}

func TestPostServiceUpdatePostReturnsNotFoundForDifferentOwner(t *testing.T) {
	t.Parallel()

	updateCalled := false
	service := postservice.NewPostService(newTestConfig(), &postRepositoryStub{
		getPostByIDFunc: func(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
			return &model.PostWithUserModel{
				ID:     postID,
				UserID: 99,
			}, nil
		},
		updatePostFunc: func(ctx context.Context, post *model.PostModel, postID int) error {
			updateCalled = true
			return nil
		},
	}, &commentRepositoryStub{})

	err := service.UpdatePost(context.Background(), &dto.CreateOrUpdatePostRequest{
		Title:   "updated",
		Content: "updated content",
	}, 10, 7)

	assertStatusCode(t, err, http.StatusNotFound)

	if err.Error() != "tweet not found" {
		t.Fatalf("expected tweet not found error, got %q", err.Error())
	}

	if updateCalled {
		t.Fatal("expected update post not to be called")
	}
}

func TestPostServiceDeletePostSuccess(t *testing.T) {
	t.Parallel()

	var deletedPostID int
	service := postservice.NewPostService(newTestConfig(), &postRepositoryStub{
		getPostByIDFunc: func(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
			return &model.PostWithUserModel{
				ID:     postID,
				UserID: 7,
			}, nil
		},
		softDeletePostFunc: func(ctx context.Context, postID int, now time.Time) error {
			deletedPostID = postID
			return nil
		},
	}, &commentRepositoryStub{})

	err := service.DeletePost(context.Background(), 10, 7)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if deletedPostID != 10 {
		t.Fatalf("expected deleted post id 10, got %d", deletedPostID)
	}
}

func TestPostServiceLikeOrUnlikePostDeletesExistingLike(t *testing.T) {
	t.Parallel()

	deleteCalled := false
	storeCalled := false
	service := postservice.NewPostService(newTestConfig(), &postRepositoryStub{
		getPostByIDFunc: func(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
			return &model.PostWithUserModel{
				ID: postID,
			}, nil
		},
		isUserAlreadyLikeFunc: func(ctx context.Context, postID, userID int) (bool, error) {
			return true, nil
		},
		deleteLikePostFunc: func(ctx context.Context, postID, userID int) error {
			deleteCalled = true
			return nil
		},
		storeLikePostFunc: func(ctx context.Context, like *model.PostLikeModel) error {
			storeCalled = true
			return nil
		},
	}, &commentRepositoryStub{})

	err := service.LikeOrUnlikePost(context.Background(), 11, 4)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected delete like to be called")
	}

	if storeCalled {
		t.Fatal("expected store like not to be called")
	}
}

func TestPostServiceLikeOrUnlikePostReturnsInternalErrorWhenCheckFails(t *testing.T) {
	t.Parallel()

	service := postservice.NewPostService(newTestConfig(), &postRepositoryStub{
		getPostByIDFunc: func(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
			return &model.PostWithUserModel{
				ID: postID,
			}, nil
		},
		isUserAlreadyLikeFunc: func(ctx context.Context, postID, userID int) (bool, error) {
			return false, errors.New("failed")
		},
	}, &commentRepositoryStub{})

	err := service.LikeOrUnlikePost(context.Background(), 11, 4)

	assertStatusCode(t, err, http.StatusInternalServerError)

	if err.Error() != "failed to check post like" {
		t.Fatalf("expected failed to check post like error, got %q", err.Error())
	}
}
