package server

import (
	"closealerts/app/handlers"
	"closealerts/app/types"
	"context"

	"go.uber.org/fx"
)

func RegisterListeningWebhooks(lc fx.Lifecycle, config types.Config, upd handlers.UpdateHandler) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case update := <-config.Updates:
						upd.Handle(ctx, update)
					}
				}
			}()

			return nil
		},
	})
}
