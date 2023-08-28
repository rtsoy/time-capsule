package handler

import (
	"fmt"
	"net/http"

	"time-capsule/internal/service"

	"github.com/julienschmidt/httprouter"
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
	h.router.GET("/", h.Index)
}

func (h *handler) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "hello!")
}
