package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"time-capsule/internal/domain"
	"time-capsule/internal/service"
	mock_service "time-capsule/internal/service/mocks"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"
)

func TestAuthHandler_signIn(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUserService, ctx context.Context, email, password string)

	tests := []struct {
		name                 string
		mockBehavior         mockBehavior
		inputBody            string
		inputData            domain.LogInUserDTO
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			mockBehavior: func(s *mock_service.MockUserService, ctx context.Context, email, password string) {
				s.EXPECT().GenerateToken(ctx, email, password).Return("some-token", nil).Times(1)
			},
			inputBody: `{"email": "foo@example.com", "password": "Qwerty123"}`,
			inputData: domain.LogInUserDTO{
				Email:    "foo@example.com",
				Password: "Qwerty123",
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"token":"some-token"}`,
		},
		{
			name:                 "Invalid JSON",
			mockBehavior:         func(s *mock_service.MockUserService, ctx context.Context, email, password string) {},
			inputBody:            `{`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid json"}`,
		},
		{
			name: "Service-Failure",
			mockBehavior: func(s *mock_service.MockUserService, ctx context.Context, email, password string) {
				s.EXPECT().GenerateToken(ctx, email, password).Return("", errors.New("some error")).Times(1)
			},
			inputBody: `{"email": "foo@example.com", "password": "Qwerty123"}`,
			inputData: domain.LogInUserDTO{
				Email:    "foo@example.com",
				Password: "Qwerty123",
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"some error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				ctx = context.Background()

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

			test.mockBehavior(userSvc, ctx, test.inputData.Email, test.inputData.Password)

			router.POST(signInURL, hndlr.signIn)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, signInURL, bytes.NewBufferString(test.inputBody))
			req.Header.Add("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestAuthHandler_signUp(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUserService, ctx context.Context, input domain.CreateUserDTO)

	tests := []struct {
		name                 string
		mockBehavior         mockBehavior
		inputBody            string
		inputData            domain.CreateUserDTO
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			mockBehavior: func(s *mock_service.MockUserService, ctx context.Context, input domain.CreateUserDTO) {
				s.EXPECT().CreateUser(ctx, input).Return(&domain.User{
					ID:           primitive.NilObjectID,
					Username:     input.Username,
					Email:        input.Email,
					RegisteredAt: time.Unix(0, 0),
				}, nil).Times(1)
			},
			inputBody: `{"username": "username123", "email": "foo@example.com", "password": "Qwerty123"}`,
			inputData: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "Qwerty123",
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"000000000000000000000000","username":"username123","email":"foo@example.com","registeredAt":"1970-01-01T00:00:00Z"}`,
		},
		{
			name:                 "Invalid JSON",
			mockBehavior:         func(s *mock_service.MockUserService, ctx context.Context, input domain.CreateUserDTO) {},
			inputBody:            `{`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid json"}`,
		},
		{
			name: "Service-Failure",
			mockBehavior: func(s *mock_service.MockUserService, ctx context.Context, input domain.CreateUserDTO) {
				s.EXPECT().CreateUser(ctx, input).Return(nil, errors.New("some error")).Times(1)
			},
			inputBody: `{"username": "username123", "email": "foo@example.com", "password": "Qwerty123"}`,
			inputData: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "Qwerty123",
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"some error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				ctx = context.Background()

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

			test.mockBehavior(userSvc, ctx, test.inputData)

			router.POST(signUpURL, hndlr.signUp)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, signUpURL, bytes.NewBufferString(test.inputBody))
			req.Header.Add("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}
