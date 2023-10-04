package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCapsuleHandler_deleteCapsule(t *testing.T) {
	type mockBehavior func(s *mock_service.MockCapsuleService, ctx context.Context,
		userID, capsuleID primitive.ObjectID)

	tests := []struct {
		name                 string
		mockBehavior         mockBehavior
		ctxUserID            string
		capsuleID            primitive.ObjectID
		capsuleIDHex         string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {
				s.EXPECT().DeleteCapsule(ctx, userID, capsuleID).Return(nil).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusNoContent,
			expectedResponseBody: "",
		},
		{
			name:                 "Invalid-Context",
			mockBehavior:         func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {},
			ctxUserID:            "123123123",
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name:                 "Invalid-CapsuleID",
			mockBehavior:         func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         "1321231232",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid id"}`,
		},
		{
			name: "Service-Failure",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {
				s.EXPECT().DeleteCapsule(ctx, userID, capsuleID).Return(errors.New("some error")).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"some error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Run(test.name, func(t *testing.T) {
				c := gomock.NewController(t)
				defer c.Finish()

				var (
					ctx = context.WithValue(context.Background(), userCtx, test.ctxUserID)

					capSvc = mock_service.NewMockCapsuleService(c)
					svc    = &service.Service{
						CapsuleService: capSvc,
					}
					router = httprouter.New()

					hndlr = handler{
						router:  router,
						svc:     svc,
						storage: nil,
					}
				)

				test.mockBehavior(capSvc, ctx, primitive.NilObjectID, test.capsuleID)

				router.DELETE(deleteCapsule, hndlr.deleteCapsule)

				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodDelete, getCapsulesURL+"/"+test.capsuleIDHex, nil)
				req = req.WithContext(ctx)

				router.ServeHTTP(w, req)

				assert.Equal(t, test.expectedStatusCode, w.Code)
				assert.Equal(t, test.expectedResponseBody, w.Body.String())
			})
		})
	}
}

func TestCapsuleHandler_updateCapsule(t *testing.T) {
	type mockBehavior func(s *mock_service.MockCapsuleService, ctx context.Context,
		userID, capsuleID primitive.ObjectID, update domain.UpdateCapsuleDTO)

	tests := []struct {
		name                 string
		mockBehavior         mockBehavior
		ctxUserID            string
		capsuleID            primitive.ObjectID
		capsuleIDHex         string
		inputBody            string
		inputData            domain.UpdateCapsuleDTO
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				s.EXPECT().UpdateCapsule(ctx, userID, capsuleID, gomock.Any()).Return(nil).Times(1)
			},
			ctxUserID:    primitive.NilObjectID.Hex(),
			capsuleID:    primitive.NilObjectID,
			capsuleIDHex: primitive.NilObjectID.Hex(),
			inputBody:    `{"message": "brand new message", "openAt": "1970-01-01T00:00:00Z"}`,
			inputData: domain.UpdateCapsuleDTO{
				Message: "brand new message",
				OpenAt:  time.Unix(0, 0),
			},
			expectedStatusCode:   http.StatusNoContent,
			expectedResponseBody: "",
		},
		{
			name: "Invalid-Context",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, update domain.UpdateCapsuleDTO) {
			},
			ctxUserID:            "12312312",
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			inputBody:            `{"message": "brand new message", "openAt": "1970-01-01T00:00:00Z"}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name: "Invalid-CapsuleID",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, update domain.UpdateCapsuleDTO) {
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         "123123213",
			inputBody:            `{"message": "brand new message", "openAt": "1970-01-01T00:00:00Z"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid id"}`,
		},
		{
			name: "Invalid-JSON",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, update domain.UpdateCapsuleDTO) {
			},
			ctxUserID:    primitive.NilObjectID.Hex(),
			capsuleID:    primitive.NilObjectID,
			capsuleIDHex: primitive.NilObjectID.Hex(),
			inputBody:    `{"message": 123, "openAt": "1970-01-01T00:00:00Z"}`,
			inputData: domain.UpdateCapsuleDTO{
				Message: "brand new message",
				OpenAt:  time.Unix(0, 0),
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid json"}`,
		},
		{
			name: "Service-Failure",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				s.EXPECT().UpdateCapsule(ctx, userID, capsuleID, gomock.Any()).Return(errors.New("some error")).Times(1)
			},
			ctxUserID:    primitive.NilObjectID.Hex(),
			capsuleID:    primitive.NilObjectID,
			capsuleIDHex: primitive.NilObjectID.Hex(),
			inputBody:    `{"message": "brand new message", "openAt": "1970-01-01T00:00:00Z"}`,
			inputData: domain.UpdateCapsuleDTO{
				Message: "brand new message",
				OpenAt:  time.Unix(0, 0),
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
				ctx = context.WithValue(context.Background(), userCtx, test.ctxUserID)

				capSvc = mock_service.NewMockCapsuleService(c)
				svc    = &service.Service{
					CapsuleService: capSvc,
				}
				router = httprouter.New()

				hndlr = handler{
					router:  router,
					svc:     svc,
					storage: nil,
				}
			)

			test.mockBehavior(capSvc, ctx, primitive.NilObjectID, test.capsuleID, test.inputData)

			router.PATCH(updateCapsule, hndlr.updateCapsule)

			w := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPatch, getCapsulesURL+"/"+test.capsuleIDHex, bytes.NewBufferString(test.inputBody))
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestCapsuleHandler_getCapsuleByID(t *testing.T) {
	type mockBehavior func(s *mock_service.MockCapsuleService, ctx context.Context,
		userID, capsuleID primitive.ObjectID)

	tests := []struct {
		name                 string
		mockBehavior         mockBehavior
		ctxUserID            string
		capsuleID            primitive.ObjectID
		capsuleIDHex         string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, capsuleID primitive.ObjectID) {
				s.EXPECT().GetCapsuleByID(ctx, userID, capsuleID).Return(&domain.Capsule{
					ID:        primitive.NilObjectID,
					UserID:    primitive.NilObjectID,
					Message:   "some message",
					Images:    []string{},
					OpenAt:    time.Unix(1, 0),
					CreatedAt: time.Unix(0, 0),
					Notified:  true,
				}, nil).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"000000000000000000000000","userID":"000000000000000000000000","message":"some message","images":[],"openAt":"1970-01-01T00:00:01Z","createdAt":"1970-01-01T00:00:00Z"}`,
		},
		{
			name: "Invalid-Context",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, capsuleID primitive.ObjectID) {
			},
			ctxUserID:            "12312321321",
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name: "Invalid-CapsuleID",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, capsuleID primitive.ObjectID) {
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         "12312321",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid id"}`,
		},
		{
			name: "Service-Failure",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, capsuleID primitive.ObjectID) {
				s.EXPECT().GetCapsuleByID(ctx, userID, capsuleID).Return(nil, errors.New("some error")).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"some error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				ctx = context.WithValue(context.Background(), userCtx, test.ctxUserID)

				capSvc = mock_service.NewMockCapsuleService(c)
				svc    = &service.Service{
					CapsuleService: capSvc,
				}
				router = httprouter.New()

				hndlr = handler{
					router:  router,
					svc:     svc,
					storage: nil,
				}
			)

			test.mockBehavior(capSvc, ctx, primitive.NilObjectID, test.capsuleID)

			router.GET(getCapsuleURL, hndlr.getCapsuleByID)

			w := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, getCapsulesURL+"/"+test.capsuleIDHex, nil)
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestCapsuleHandler_getCapsules(t *testing.T) {
	type mockBehavior func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID)

	tests := []struct {
		name                 string
		mockBehavior         mockBehavior
		ctxUserID            string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID) {
				s.EXPECT().GetAllCapsules(ctx, userID).
					Return([]*domain.Capsule{
						{
							ID:        primitive.NilObjectID,
							UserID:    primitive.NilObjectID,
							Message:   "some message 1",
							Images:    []string{},
							OpenAt:    time.Unix(1, 0),
							CreatedAt: time.Unix(0, 0),
							Notified:  false,
						},
						{
							ID:        primitive.NilObjectID,
							UserID:    primitive.NilObjectID,
							Message:   "some message 2",
							Images:    []string{},
							OpenAt:    time.Unix(2, 0),
							CreatedAt: time.Unix(3, 0),
							Notified:  true,
						},
					}, nil).Times(1)
			},
			ctxUserID:          primitive.NilObjectID.Hex(),
			expectedStatusCode: http.StatusOK,
			expectedResponseBody: strings.Replace(strings.Replace(`[
					{"id":"000000000000000000000000","userID":"000000000000000000000000","message":"some message 1","images":[],"openAt":"1970-01-01T00:00:01Z","createdAt":"1970-01-01T00:00:00Z"},
					{"id":"000000000000000000000000","userID":"000000000000000000000000","message":"some message 2","images":[],"openAt":"1970-01-01T00:00:02Z","createdAt":"1970-01-01T00:00:03Z"}
			]`, "\n", "", -1), "\t", "", -1),
		},
		{
			name:                 "Invalid-Context",
			mockBehavior:         func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID) {},
			ctxUserID:            "12321312",
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name: "Service-Failure",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID) {
				s.EXPECT().GetAllCapsules(ctx, userID).
					Return(nil, errors.New("some error")).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"some error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				ctx = context.WithValue(context.Background(), userCtx, test.ctxUserID)

				capSvc = mock_service.NewMockCapsuleService(c)
				svc    = &service.Service{
					CapsuleService: capSvc,
				}
				router = httprouter.New()

				hndlr = handler{
					router:  router,
					svc:     svc,
					storage: nil,
				}
			)

			test.mockBehavior(capSvc, ctx, primitive.NilObjectID)

			router.GET(getCapsulesURL, hndlr.getCapsules)

			w := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, getCapsulesURL, nil)
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestCapsuleHandler_createCapsule(t *testing.T) {
	type mockBehavior func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO)

	tests := []struct {
		name                 string
		mockBehavior         mockBehavior
		ctxUserID            string
		inputBody            string
		inputData            domain.CreateCapsuleDTO
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
				s.EXPECT().CreateCapsule(ctx, userID, gomock.Any()).Return(
					&domain.Capsule{
						ID:        primitive.NilObjectID,
						UserID:    primitive.NilObjectID,
						Message:   "some message",
						Images:    []string{},
						OpenAt:    time.Unix(0, 0).UTC(),
						CreatedAt: time.Unix(0, 0).UTC(),
						Notified:  false,
					}, nil).Times(1)
			},
			ctxUserID: primitive.NilObjectID.Hex(),
			inputBody: `{"message":"some message", "openAt": "1970-01-01T00:00:00Z"}`,
			inputData: domain.CreateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Unix(0, 0),
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"000000000000000000000000","userID":"000000000000000000000000","message":"some message","images":[],"openAt":"1970-01-01T00:00:00Z","createdAt":"1970-01-01T00:00:00Z"}`,
		},
		{
			name: "Invalid-Context",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
			},
			ctxUserID:            "12312321312",
			inputBody:            `{"message":"some message", "openAt": "1970-01-01T00:00:00Z"}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name: "Invalid-JSON",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			inputBody:            `{"message":"some message", "openAt": "1970-01-01T06:0"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid timestamp"}`,
		},
		{
			name: "Service-Failure",
			mockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
				s.EXPECT().CreateCapsule(ctx, userID, gomock.Any()).Return(nil, errors.New("some error")).Times(1)
			},
			ctxUserID: primitive.NilObjectID.Hex(),
			inputBody: `{"message":"some message", "openAt": "1970-01-01T00:00:00Z"}`,
			inputData: domain.CreateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Unix(0, 0),
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
				ctx = context.WithValue(context.Background(), userCtx, test.ctxUserID)

				capSvc = mock_service.NewMockCapsuleService(c)
				svc    = &service.Service{
					CapsuleService: capSvc,
				}
				router = httprouter.New()

				hndlr = handler{
					router:  router,
					svc:     svc,
					storage: nil,
				}
			)

			test.mockBehavior(capSvc, ctx, primitive.NilObjectID, test.inputData)

			router.POST(createCapsuleURL, hndlr.createCapsule)

			w := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPost, createCapsuleURL, bytes.NewBufferString(test.inputBody))
			req = req.WithContext(ctx)
			req.Header.Add("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}
