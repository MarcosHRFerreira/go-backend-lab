package unit_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	commentservice "go-tweets/internal/service/comment"
)

func TestCommentServiceCreateCommentSuccess(t *testing.T) {
	t.Parallel()

	var storedComment *model.CommentModel
	service := commentservice.NewCommentService(newTestConfig(), &commentRepositoryStub{
		storeCommentFunc: func(ctx context.Context, comment *model.CommentModel) error {
			storedComment = comment
			return nil
		},
	}, &postRepositoryStub{
		getPostByIDFunc: func(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
			return &model.PostWithUserModel{
				ID: postID,
			}, nil
		},
	})

	err := service.CreateComment(context.Background(), &dto.StoreCommentRequest{
		PostID:  12,
		Content: "great post",
	}, 7)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if storedComment == nil {
		t.Fatal("expected comment to be stored")
	}

	if storedComment.PostID != 12 || storedComment.UserID != 7 || storedComment.Content != "great post" {
		t.Fatalf("unexpected stored comment payload: %+v", storedComment)
	}
}

func TestCommentServiceCreateCommentReturnsNotFoundWhenPostDoesNotExist(t *testing.T) {
	t.Parallel()

	storeCalled := false
	service := commentservice.NewCommentService(newTestConfig(), &commentRepositoryStub{
		storeCommentFunc: func(ctx context.Context, comment *model.CommentModel) error {
			storeCalled = true
			return nil
		},
	}, &postRepositoryStub{
		getPostByIDFunc: func(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
			return nil, nil
		},
	})

	err := service.CreateComment(context.Background(), &dto.StoreCommentRequest{
		PostID:  12,
		Content: "great post",
	}, 7)

	assertStatusCode(t, err, http.StatusNotFound)

	if err.Error() != "tweet not found" {
		t.Fatalf("expected tweet not found error, got %q", err.Error())
	}

	if storeCalled {
		t.Fatal("expected store comment not to be called")
	}
}

func TestCommentServiceLikeOrUnlikeCommentDeletesExistingLike(t *testing.T) {
	t.Parallel()

	deleteCalled := false
	storeCalled := false
	service := commentservice.NewCommentService(newTestConfig(), &commentRepositoryStub{
		detailCommentFunc: func(ctx context.Context, commentID int) (*model.CommentModel, error) {
			return &model.CommentModel{ID: commentID}, nil
		},
		isUserAlreadyLikeFunc: func(ctx context.Context, commentID, userID int) (bool, error) {
			return true, nil
		},
		deleteLikeCommentFunc: func(ctx context.Context, commentID, userID int) error {
			deleteCalled = true
			return nil
		},
		storeLikeCommentFunc: func(ctx context.Context, like *model.CommentLikeModel) error {
			storeCalled = true
			return nil
		},
	}, &postRepositoryStub{})

	err := service.LikeOrUnLikeComment(context.Background(), 4, 9)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected delete like comment to be called")
	}

	if storeCalled {
		t.Fatal("expected store like comment not to be called")
	}
}

func TestCommentServiceLikeOrUnlikeCommentReturnsNotFoundWhenCommentDoesNotExist(t *testing.T) {
	t.Parallel()

	service := commentservice.NewCommentService(newTestConfig(), &commentRepositoryStub{
		detailCommentFunc: func(ctx context.Context, commentID int) (*model.CommentModel, error) {
			return nil, nil
		},
	}, &postRepositoryStub{})

	err := service.LikeOrUnLikeComment(context.Background(), 4, 9)

	assertStatusCode(t, err, http.StatusNotFound)

	if err.Error() != "comment not found" {
		t.Fatalf("expected comment not found error, got %q", err.Error())
	}
}

func TestCommentServiceLikeOrUnlikeCommentReturnsInternalErrorWhenDeleteFails(t *testing.T) {
	t.Parallel()

	service := commentservice.NewCommentService(newTestConfig(), &commentRepositoryStub{
		detailCommentFunc: func(ctx context.Context, commentID int) (*model.CommentModel, error) {
			return &model.CommentModel{ID: commentID}, nil
		},
		isUserAlreadyLikeFunc: func(ctx context.Context, commentID, userID int) (bool, error) {
			return true, nil
		},
		deleteLikeCommentFunc: func(ctx context.Context, commentID, userID int) error {
			return errors.New("delete failed")
		},
	}, &postRepositoryStub{})

	err := service.LikeOrUnLikeComment(context.Background(), 4, 9)

	assertStatusCode(t, err, http.StatusInternalServerError)

	if err.Error() != "failed to delete comment like" {
		t.Fatalf("expected failed to delete comment like error, got %q", err.Error())
	}
}
