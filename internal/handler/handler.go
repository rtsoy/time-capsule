package handler

import (
	"errors"
	"net/http"

	"time-capsule/internal/service"

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
	getCapsuleByIDURL = getCapsulesURL + "/:" + pathCapsuleID
)

type Handler interface {
	Router() http.Handler
	InitRoutes()
}

type handler struct {
	router *httprouter.Router
	svc    *service.Service
}

func NewHandler(svc *service.Service) Handler {
	router := httprouter.New()

	return &handler{
		router: router,
		svc:    svc,
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
	h.router.GET(getCapsuleByIDURL, h.JWTAuthentication(h.getCapsuleByID))
}

func parseObjectIDFromParam(params httprouter.Params, name string) (primitive.ObjectID, error) {
	id := params.ByName(name)

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid id")
	}

	return oid, nil
}
