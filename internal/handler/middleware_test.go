package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"time-capsule/internal/service"
	mock_service "time-capsule/internal/service/mocks"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"
)

func TestMiddlewareHandler_RateLimiter(t *testing.T) {
	tests := []struct {
		name                 string
		requests             int
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                 "OK",
			requests:             requestRateLimit - 1,
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "ok",
		},
		{
			name:                 "Rate Limit",
			requests:             requestRateLimit + 1,
			expectedStatusCode:   http.StatusTooManyRequests,
			expectedResponseBody: `{"message":"rate limit exceeded"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				router = httprouter.New()

				hndlr = &handler{
					router:  router,
					svc:     nil,
					storage: nil,
				}
			)

			router.GET("/", hndlr.RateLimiter(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "ok")
			}))

			var w *httptest.ResponseRecorder
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			for i := 0; i < test.requests; i++ {
				w = httptest.NewRecorder()

				router.ServeHTTP(w, req)
			}

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestMiddlewareHandler_JWTAuthentication(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUserService, token string)

	tests := []struct {
		name                 string
		mockBehavior         mockBehavior
		headerName           string
		headerValue          string
		token                string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			mockBehavior: func(s *mock_service.MockUserService, token string) {
				s.EXPECT().ParseToken(token).Return(jwt.MapClaims{
					"userID": primitive.NilObjectID.Hex(),
					"exp":    time.Now().Add(1 * time.Hour).Unix(),
				}, nil).Times(1)
			},
			headerName:           "Authorization",
			headerValue:          "Bearer token",
			token:                "token",
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: primitive.NilObjectID.Hex(),
		},
		{
			name:                 "Empty-Auth-Header",
			mockBehavior:         func(s *mock_service.MockUserService, token string) {},
			headerName:           "",
			headerValue:          "",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"empty auth header"}`,
		},
		{
			name:                 "Invalid-Auth-Header",
			mockBehavior:         func(s *mock_service.MockUserService, token string) {},
			headerName:           "Authorization",
			headerValue:          "Bumblebee token",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"invalid auth header"}`,
		},
		{
			name:                 "Empty-Token",
			mockBehavior:         func(s *mock_service.MockUserService, token string) {},
			headerName:           "Authorization",
			headerValue:          "Bearer ",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"token is empty"}`,
		},
		{
			name: "Service-Failure",
			mockBehavior: func(s *mock_service.MockUserService, token string) {
				s.EXPECT().ParseToken(token).Return(nil, errors.New("some error")).Times(1)
			},
			headerName:           "Authorization",
			headerValue:          "Bearer token",
			token:                "token",
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"some error"}`,
		},
		{
			name: "No-UserID-In-Claims",
			mockBehavior: func(s *mock_service.MockUserService, token string) {
				s.EXPECT().ParseToken(token).Return(jwt.MapClaims{
					"exp": time.Now().Add(1 * time.Hour).Unix(),
				}, nil).Times(1)
			},
			headerName:           "Authorization",
			headerValue:          "Bearer token",
			token:                "token",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"invalid token"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				userSvc = mock_service.NewMockUserService(c)
				svc     = &service.Service{
					UserService: userSvc,
				}
				router = httprouter.New()

				hndlr = handler{
					router:  router,
					svc:     svc,
					storage: nil,
				}
			)

			test.mockBehavior(userSvc, test.token)

			router.GET("/", hndlr.JWTAuthentication(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
				oid, _ := getUserID(r)
				fmt.Fprint(w, oid.Hex())
			}))

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Add(test.headerName, test.headerValue)

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestMiddlewareHandler_getUserID(t *testing.T) {
	tests := []struct {
		name          string
		request       func() *http.Request
		userID        primitive.ObjectID
		expectedError error
	}{
		{
			name: "OK",
			request: func() *http.Request {
				r := &http.Request{}

				return r.WithContext(context.WithValue(r.Context(), userCtx, primitive.NilObjectID.Hex()))
			},
			userID:        primitive.NilObjectID,
			expectedError: nil,
		},
		{
			name: "Conversion-Failure",
			request: func() *http.Request {
				r := &http.Request{}
				return r
			},
			userID:        primitive.NilObjectID,
			expectedError: errors.New("internal server error"),
		},
		{
			name: "Empty-Context-UserID",
			request: func() *http.Request {
				r := &http.Request{}

				return r.WithContext(context.WithValue(r.Context(), userCtx, ""))
			},
			userID:        primitive.NilObjectID,
			expectedError: errors.New("internal server error"),
		},
		{
			name: "Not-ObjectID",
			request: func() *http.Request {
				r := &http.Request{}

				return r.WithContext(context.WithValue(r.Context(), userCtx, "some_id"))
			},
			userID:        primitive.NilObjectID,
			expectedError: errors.New("internal server error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oid, err := getUserID(test.request())

			if err != nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.Equal(t, oid, test.userID)
			}
		})
	}
}
