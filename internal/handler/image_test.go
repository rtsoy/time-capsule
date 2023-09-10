package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"time-capsule/internal/domain"
	"time-capsule/internal/service"
	mock_service "time-capsule/internal/service/mocks"
	mock_storage "time-capsule/internal/storage/mocks"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"
)

func TestImageHandler_removeCapsuleImage(t *testing.T) {
	type serviceMockBehavior func(s *mock_service.MockCapsuleService, ctx context.Context,
		userID, capsuleID primitive.ObjectID, imageID string)

	type storageMockBehavior func(s *mock_storage.MockStorage, ctx context.Context,
		fileName string)

	tests := []struct {
		name                 string
		serviceMockBehavior  serviceMockBehavior
		storageMockBehavior  storageMockBehavior
		ctxUserID            string
		capsuleID            primitive.ObjectID
		capsuleIDHex         string
		imageID              primitive.ObjectID
		imageIDHex           string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, imageID string) {
				s.EXPECT().RemoveImage(ctx, userID, capsuleID, imageID).Return(nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {
				s.EXPECT().Delete(ctx, fileName).Return(nil).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusNoContent,
			expectedResponseBody: "",
		},
		{
			name: "Invalid-Context",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, imageID string) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {},
			ctxUserID:            "123123123",
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name: "Invalid-CapsuleID",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, imageID string) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         "12321321312",
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid id"}`,
		},
		{
			name: "Invalid-ImageID",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, imageID string) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           "12312321321",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid id"}`,
		},
		{
			name: "Service-Failure",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, imageID string) {
				s.EXPECT().RemoveImage(ctx, userID, capsuleID, imageID).Return(errors.New("some error")).Times(1)
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"some error"}`,
		},
		{
			name: "Storage-Failure",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, imageID string) {
				s.EXPECT().RemoveImage(ctx, userID, capsuleID, imageID).Return(nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {
				s.EXPECT().Delete(ctx, fileName).Return(errors.New("some error")).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
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

				strge  = mock_storage.NewMockStorage(c)
				router = httprouter.New()

				hndlr = handler{
					router:  router,
					svc:     svc,
					storage: strge,
				}
			)

			test.serviceMockBehavior(capSvc, ctx, primitive.NilObjectID, test.capsuleID, test.imageIDHex)
			test.storageMockBehavior(strge, ctx, test.imageIDHex)

			router.DELETE(removeCapsuleImage, hndlr.removeCapsuleImage)

			w := httptest.NewRecorder()

			targetURL := strings.Replace(getCapsuleImage, ":capsuleID", test.capsuleIDHex, 1)
			targetURL = strings.Replace(targetURL, ":imageID", test.imageIDHex, 1)

			req := httptest.NewRequest(http.MethodDelete, targetURL, nil)
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestImageHandler_getCapsuleImage(t *testing.T) {
	type serviceMockBehavior func(s *mock_service.MockCapsuleService, ctx context.Context,
		userID, capsuleID primitive.ObjectID)

	type storageMockBehavior func(s *mock_storage.MockStorage, ctx context.Context,
		fileName string)

	tests := []struct {
		name                 string
		serviceMockBehavior  serviceMockBehavior
		storageMockBehavior  storageMockBehavior
		ctxUserID            string
		capsuleID            primitive.ObjectID
		capsuleIDHex         string
		imageID              primitive.ObjectID
		imageIDHex           string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {
				s.EXPECT().GetCapsuleByID(ctx, userID, capsuleID).Return(nil, nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {
				s.EXPECT().Get(ctx, fileName).Return(&domain.File{
					Bytes: []byte("good"),
					Name:  "test-file",
					Size:  12312,
				}, nil).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "good",
		},
		{
			name: "Invalid-Context",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {},
			ctxUserID:            "123213213",
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name:                "Invalid-CapsuleID",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         "123123213",
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid id"}`,
		},
		{
			name: "Invalid-ImageID",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {

			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           "12321312312",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid id"}`,
		},
		{
			name: "Service-Failure",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {
				s.EXPECT().GetCapsuleByID(ctx, userID, capsuleID).Return(nil, errors.New("some error")).Times(1)
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"some error"}`,
		},
		{
			name: "Storage-Failure",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID) {
				s.EXPECT().GetCapsuleByID(ctx, userID, capsuleID).Return(nil, nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, fileName string) {
				s.EXPECT().Get(ctx, fileName).Return(nil, errors.New("some error")).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			imageID:              primitive.NilObjectID,
			imageIDHex:           primitive.NilObjectID.Hex(),
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"not found"}`,
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

				strge  = mock_storage.NewMockStorage(c)
				router = httprouter.New()

				hndlr = handler{
					router:  router,
					svc:     svc,
					storage: strge,
				}
			)

			test.serviceMockBehavior(capSvc, ctx, primitive.NilObjectID, test.capsuleID)
			test.storageMockBehavior(strge, ctx, test.imageIDHex)

			router.GET(getCapsuleImage, hndlr.getCapsuleImage)

			w := httptest.NewRecorder()

			targetURL := strings.Replace(getCapsuleImage, ":capsuleID", test.capsuleIDHex, 1)
			targetURL = strings.Replace(targetURL, ":imageID", test.imageIDHex, 1)

			req := httptest.NewRequest(http.MethodGet, targetURL, nil)
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestImageHandler_addCapsuleImage(t *testing.T) {
	type serviceMockBehavior func(s *mock_service.MockCapsuleService, ctx context.Context,
		userID, capsuleID primitive.ObjectID, image string)

	type storageMockBehavior func(s *mock_storage.MockStorage, ctx context.Context,
		file domain.File)

	oidPatch := gomonkey.ApplyFunc(primitive.NewObjectID, func() primitive.ObjectID { return primitive.NilObjectID })
	defer oidPatch.Reset()

	tests := []struct {
		name                 string
		serviceMockBehavior  serviceMockBehavior
		storageMockBehavior  storageMockBehavior
		ctxUserID            string
		capsuleID            primitive.ObjectID
		capsuleIDHex         string
		inputData            domain.File
		uploadInput          string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK-PNG",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
				s.EXPECT().AddImage(ctx, userID, capsuleID, image).Return(nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {
				s.EXPECT().Upload(ctx, file).Return(nil).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			uploadInput:          "./fixtures/images/ok.png",
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"name":"000000000000000000000000.png","size":21656}`,
		},
		{
			name: "OK-JPEG",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
				s.EXPECT().AddImage(ctx, userID, capsuleID, image).Return(nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {
				s.EXPECT().Upload(ctx, file).Return(nil).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			uploadInput:          "./fixtures/images/ok.jpg",
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"name":"000000000000000000000000.jpg","size":14327}`,
		},
		{
			name: "Invalid-Context",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {},
			ctxUserID:            "12312312312",
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			uploadInput:          "./fixtures/images/ok.png",
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name: "Invalid-CapsuleID",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         "12321312312312",
			uploadInput:          "./fixtures/images/ok.png",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid id"}`,
		},
		{
			name: "File-Too-Large",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			uploadInput:          "./fixtures/images/large.jpg",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"unable to parse the form"}`,
		},
		{
			name: "File-Reading-Failure",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			uploadInput:          "./fixtures/images/unreadable",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"failed to read the uploaded file"}`,
		},
		{
			name: "File-Wrong-Type",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
			},
			storageMockBehavior:  func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			uploadInput:          "./fixtures/images/wrong.gif",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid file type"}`,
		},
		{
			name: "Storage-Failure",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {
				s.EXPECT().Upload(ctx, file).Return(errors.New("some error")).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			uploadInput:          "./fixtures/images/ok.png",
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
		{
			name: "Service-Failure",
			serviceMockBehavior: func(s *mock_service.MockCapsuleService, ctx context.Context, userID, capsuleID primitive.ObjectID, image string) {
				s.EXPECT().AddImage(ctx, userID, capsuleID, image).Return(errors.New("some error")).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, file domain.File) {
				s.EXPECT().Upload(ctx, file).Return(nil).Times(1)
			},
			ctxUserID:            primitive.NilObjectID.Hex(),
			capsuleID:            primitive.NilObjectID,
			capsuleIDHex:         primitive.NilObjectID.Hex(),
			uploadInput:          "./fixtures/images/ok.png",
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			file, err := os.Open(test.uploadInput)
			assert.NoError(t, err)
			defer file.Close()

			stat, err := file.Stat()
			assert.NoError(t, err)

			fileBytes := make([]byte, stat.Size())

			_, err = file.Read(fileBytes)
			assert.NoError(t, err)

			ext := path.Ext(stat.Name())

			test.inputData.Bytes = fileBytes
			test.inputData.Size = stat.Size()
			test.inputData.Name = primitive.NewObjectID().Hex() + ext

			var (
				ctx = context.WithValue(context.Background(), userCtx, test.ctxUserID)

				capSvc = mock_service.NewMockCapsuleService(c)
				svc    = &service.Service{
					CapsuleService: capSvc,
				}

				strge  = mock_storage.NewMockStorage(c)
				router = httprouter.New()

				hndlr = handler{
					router:  router,
					svc:     svc,
					storage: strge,
				}
			)

			test.serviceMockBehavior(capSvc, ctx, primitive.NilObjectID, test.capsuleID, test.inputData.Name)
			test.storageMockBehavior(strge, ctx, test.inputData)

			router.POST(addCapsuleImage, hndlr.addCapsuleImage)

			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("image", test.uploadInput)
			assert.NoError(t, err)

			_, err = part.Write(fileBytes)
			assert.NoError(t, err)

			err = writer.Close()
			assert.NoError(t, err)

			w := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPost, strings.Replace(addCapsuleImage, ":capsuleID", test.capsuleIDHex, 1), body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}
