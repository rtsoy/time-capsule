package httpserver

import (
	"context"
	"net/http"
	"time"

	"time-capsule/config"
)

const (
	defaultReadTimeout    = 5 * time.Second
	defaultWriteTimeout   = 5 * time.Second
	defaultMaxHeaderBytes = 1 << 20
)

type Server struct {
	httpServer *http.Server
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Run(cfg *config.Config, handler http.Handler) error {
	httpServer := &http.Server{
		Addr:           ":" + cfg.HttpAddr,
		Handler:        handler,
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		MaxHeaderBytes: defaultMaxHeaderBytes,
	}

	return httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
