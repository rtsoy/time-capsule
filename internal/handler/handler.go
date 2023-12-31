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
}

type handler struct {
	router  *httprouter.Router
	svc     *service.Service
	storage storage.Storage
}

func NewHandler(svc *service.Service, storage storage.Storage) Handler {
	router := httprouter.New()

	h := &handler{
		router:  router,
		svc:     svc,
		storage: storage,
	}

	h.initRoutes()

	return h
}

func (h *handler) Router() http.Handler {
	return h.router
}

func (h *handler) initRoutes() {
	h.router.ServeFiles("/swagger/*filepath", http.Dir("docs"))

	h.router.POST(signUpURL, h.RateLimiter(h.signUp))
	h.router.POST(signInURL, h.RateLimiter(h.signIn))

	h.router.POST(createCapsuleURL, h.RateLimiter(h.JWTAuthentication(h.createCapsule)))
	h.router.GET(getCapsulesURL, h.RateLimiter(h.JWTAuthentication(h.getCapsules)))
	h.router.GET(getCapsuleURL, h.RateLimiter(h.JWTAuthentication(h.getCapsuleByID)))
	h.router.PATCH(updateCapsule, h.RateLimiter(h.JWTAuthentication(h.updateCapsule)))
	h.router.DELETE(deleteCapsule, h.RateLimiter(h.JWTAuthentication(h.deleteCapsule)))

	h.router.POST(addCapsuleImage, h.RateLimiter(h.JWTAuthentication(h.addCapsuleImage)))
	h.router.GET(getCapsuleImage, h.RateLimiter(h.JWTAuthentication(h.getCapsuleImage)))
	h.router.DELETE(removeCapsuleImage, h.RateLimiter(h.JWTAuthentication(h.removeCapsuleImage)))
}

func parseObjectIDFromParam(params httprouter.Params, name string) (primitive.ObjectID, error) {
	id := params.ByName(name)

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid id")
	}

	return oid, nil
}
