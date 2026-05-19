package integration_test

import (
	"context"
	"net/http"
	"testing"

	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
)

func TestCommentCreateEndpointReturnsCreatedWhenAuthorized(t *testing.T) {
	t.Parallel()

	var receivedUserID int
	var receivedRequest *dto.StoreCommentRequest
	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{
		createCommentFunc: func(ctx context.Context, req *dto.StoreCommentRequest, userID int) error {
			receivedUserID = userID
			receivedRequest = req
			return nil
		},
	})

	recorder := performJSONRequest(t, router, http.MethodPost, "/comment/", map[string]interface{}{
		"post_id": 3,
		"content": "great post",
	}, map[string]string{
		"Authorization": mustCreateToken(t, 8, "marcos"),
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	if receivedUserID != 8 {
		t.Fatalf("expected userID 8, got %d", receivedUserID)
	}

	if receivedRequest == nil || receivedRequest.PostID != 3 {
		t.Fatalf("expected request to reach service, got %+v", receivedRequest)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "successfully" {
		t.Fatalf("expected successfully message, got %v", payload["message"])
	}
}

func TestCommentCreateEndpointReturnsUnauthorizedWithoutToken(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/comment/", map[string]interface{}{
		"post_id": 3,
		"content": "great post",
	}, nil)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "missing authorization token" {
		t.Fatalf("expected missing authorization token message, got %v", payload["message"])
	}
}

func TestCommentActionEndpointReturnsNotFoundForMissingComment(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{
		likeOrUnlikeCommentFunc: func(ctx context.Context, commentID, userID int) error {
			return apperror.NotFound("comment not found")
		},
	})

	recorder := performJSONRequest(t, router, http.MethodPost, "/comment/action", map[string]interface{}{
		"comment_id": 77,
	}, map[string]string{
		"Authorization": mustCreateToken(t, 8, "marcos"),
	})

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "comment not found" {
		t.Fatalf("expected comment not found message, got %v", payload["message"])
	}
}

func TestCommentActionEndpointReturnsSuccessWhenAuthorized(t *testing.T) {
	t.Parallel()

	var receivedCommentID int
	var receivedUserID int
	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{
		likeOrUnlikeCommentFunc: func(ctx context.Context, commentID, userID int) error {
			receivedCommentID = commentID
			receivedUserID = userID
			return nil
		},
	})

	recorder := performJSONRequest(t, router, http.MethodPost, "/comment/action", map[string]interface{}{
		"comment_id": 77,
	}, map[string]string{
		"Authorization": mustCreateToken(t, 8, "marcos"),
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if receivedCommentID != 77 || receivedUserID != 8 {
		t.Fatalf("expected commentID 77 and userID 8, got %d and %d", receivedCommentID, receivedUserID)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "success" {
		t.Fatalf("expected success message, got %v", payload["message"])
	}
}
