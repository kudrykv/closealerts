package server

import (
	"closealerts/app/types"
	"context"
	"fmt"
	"net/http"

	"go.uber.org/fx"
)

type Server struct {
	server http.Server
}

func NewServer(cfg types.Config, mux *http.ServeMux) *Server {
	return &Server{server: http.Server{Addr: cfg.Addr, Handler: mux}}
}

func RegisterServer(lc fx.Lifecycle, server *Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() { _ = server.server.ListenAndServe() }()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := server.server.Shutdown(ctx); err != nil {
				return fmt.Errorf("shutdown: %w", err)
			}

			return nil
		},
	})
}
