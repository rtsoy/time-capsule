package handler

import (
	"net/http"

	"time-capsule/internal/service"

	"github.com/julienschmidt/httprouter"
)

const (
	apiPrefix = "/api/v1"

	signUpURL = apiPrefix + "/sign-up"
	signInURL = apiPrefix + "/sign-in"

	createCapsuleURL = apiPrefix + "/capsules"
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
}
