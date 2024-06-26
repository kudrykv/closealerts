package server

import (
	"closealerts/app/handlers"
	"closealerts/app/types"
	"context"

	"go.uber.org/fx"
)

func RegisterListeningWebhooks(lc fx.Lifecycle, config types.Config, upd handlers.UpdateHandler) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				for {
					select {
					case <-ctx.Done():
						close(done)
						return

					case update := <-config.Updates:
						sem := make(chan struct{}, 20)

						go func() {
							sem <- struct{}{}
							upd.Handle(ctx, update)
							<-sem
						}()
					}
				}
			}()

			return nil
		},

		OnStop: func(context.Context) error {
			cancel()
			<-done

			return nil
		},
	})
}
