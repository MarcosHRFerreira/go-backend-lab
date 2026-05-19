package integration_test

import (
	"context"
	"net/http"
	"testing"

	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
)

func TestTweetsCreateEndpointReturnsCreatedWhenAuthorized(t *testing.T) {
	t.Parallel()

	var receivedUserID int
	var receivedRequest *dto.CreateOrUpdatePostRequest
	router := newTestRouter(&userServiceStub{}, &postServiceStub{
		createPostFunc: func(ctx context.Context, req *dto.CreateOrUpdatePostRequest, userID int) (int, error) {
			receivedUserID = userID
			receivedRequest = req
			return 11, nil
		},
	}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/tweets/", map[string]string{
		"title":   "new post",
		"content": "hello market",
	}, map[string]string{
		"Authorization": mustCreateToken(t, 5, "marcos"),
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	if receivedUserID != 5 {
		t.Fatalf("expected userID 5, got %d", receivedUserID)
	}

	if receivedRequest == nil || receivedRequest.Title != "new post" {
		t.Fatalf("expected request to reach service, got %+v", receivedRequest)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["id"] != float64(11) {
		t.Fatalf("expected id 11, got %v", payload["id"])
	}
}

func TestTweetsCreateEndpointReturnsUnauthorizedWithoutToken(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/tweets/", map[string]string{
		"title":   "new post",
		"content": "hello market",
	}, nil)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "missing authorization token" {
		t.Fatalf("expected missing authorization token message, got %v", payload["message"])
	}
}

func TestTweetsListEndpointReturnsBadRequestForInvalidPage(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodGet, "/tweets/?page=0&limit=2", nil, nil)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "page must be a valid integer greater than or equal to 1" {
		t.Fatalf("expected invalid page message, got %v", payload["message"])
	}
}

func TestTweetsDetailEndpointReturnsPostPayload(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{
		detailPostFunc: func(ctx context.Context, postID int) (*dto.DetailPostResponse, error) {
			return &dto.DetailPostResponse{
				ID:        postID,
				Username:  "marcos",
				Title:     "detail",
				Content:   "detail content",
				LikeCount: 3,
				CreatedAt: "2026-05-15 10:00:00 +0000 UTC",
				UpdatedAt: "2026-05-15 10:10:00 +0000 UTC",
				Comments: []dto.Comment{
					{
						ID:       1,
						Username: "ana",
						Content:  "nice",
					},
				},
			}, nil
		},
	}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodGet, "/tweets/7/detail", nil, nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["id"] != float64(7) {
		t.Fatalf("expected id 7, got %v", payload["id"])
	}

	if payload["title"] != "detail" {
		t.Fatalf("expected title detail, got %v", payload["title"])
	}
}

func TestTweetsDetailEndpointReturnsBadRequestForInvalidPostID(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{
		detailPostFunc: func(ctx context.Context, postID int) (*dto.DetailPostResponse, error) {
			return nil, apperror.NotFound("post not found")
		},
	}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodGet, "/tweets/invalid/detail", nil, nil)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "post_id must be a valid integer" {
		t.Fatalf("expected invalid post_id message, got %v", payload["message"])
	}
}
