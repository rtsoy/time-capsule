package handler

import (
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
