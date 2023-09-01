package handler

import (
	"errors"
	"net/http"

	"time-capsule/internal/service"
	"time-capsule/internal/storage"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	apiPrefix = "/api/v1"

	signUpURL = apiPrefix + "/sign-up"
	signInURL = apiPrefix + "/sign-in"

	pathCapsuleID = "capsuleID"

	createCapsuleURL = apiPrefix + "/capsules"
	getCapsulesURL
	getCapsuleURL = getCapsulesURL + "/:" + pathCapsuleID
	updateCapsule
	deleteCapsule

	pathImageID = "imageID"

	addCapsuleImage = getCapsuleURL + "/images"
	getCapsuleImage = addCapsuleImage + "/:" + pathImageID
	removeCapsuleImage
)

type Handler interface {
	Router() http.Handler
	InitRoutes()
}

type handler struct {
	router  *httprouter.Router
	svc     *service.Service
	storage storage.Storage
}

func NewHandler(svc *service.Service, storage storage.Storage) Handler {
	router := httprouter.New()

	return &handler{
		router:  router,
		svc:     svc,
		storage: storage,
	}
}

func (h *handler) Router() http.Handler {
	return h.router
}

func (h *handler) InitRoutes() {
	h.router.POST(signUpURL, h.signUp)
	h.router.POST(signInURL, h.signIn)

	h.router.POST(createCapsuleURL, h.JWTAuthentication(h.createCapsule))
	h.router.GET(getCapsulesURL, h.JWTAuthentication(h.getCapsules))
	h.router.GET(getCapsuleURL, h.JWTAuthentication(h.getCapsuleByID))
	h.router.PATCH(updateCapsule, h.JWTAuthentication(h.updateCapsule))
	h.router.DELETE(deleteCapsule, h.JWTAuthentication(h.deleteCapsule))

	h.router.POST(addCapsuleImage, h.JWTAuthentication(h.addCapsuleImage))
	h.router.GET(getCapsuleImage, h.JWTAuthentication(h.getCapsuleImage))
	h.router.DELETE(removeCapsuleImage, h.JWTAuthentication(h.removeCapsuleImage))
}

func parseObjectIDFromParam(params httprouter.Params, name string) (primitive.ObjectID, error) {
	id := params.ByName(name)

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid id")
	}

	return oid, nil
}
