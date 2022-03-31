package server

import (
	"closealerts/app/types"
	"context"
	"fmt"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	server http.Server
}

func NewServer(log *zap.SugaredLogger, cfg types.Config, mux *http.ServeMux) *Server {
	log.Infow("preparing web server", "addr", cfg.Addr)

	return &Server{server: http.Server{Addr: cfg.Addr, Handler: mux}}
}

func RegisterServer(lc fx.Lifecycle, config types.Config, server *Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if len(config.Cert) > 0 {
				go func() { _ = server.server.ListenAndServeTLS(config.Cert, config.Key) }()
			} else {
				go func() { _ = server.server.ListenAndServe() }()
			}

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
