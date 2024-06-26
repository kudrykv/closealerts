package main

import (
	"closealerts/app/clients"
	"closealerts/app/handlers"
	"closealerts/app/jobs"
	"closealerts/app/repositories"
	types2 "closealerts/app/repositories/types"
	"closealerts/app/server"
	"closealerts/app/services"
	"closealerts/app/types"
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		fx.Provide(
			types.NewConfig,

			clients.NewDBFromSQLite,
			clients.NewLogger,
			clients.NewSugaredLogger,
			clients.NewTelegram,

			repositories.NewAlerts,
			repositories.NewNotification,
			repositories.NewChats,
			repositories.NewMaps,

			services.NewFakes,
			services.NewAlerts,
			services.NewNotification,
			services.NewChats,
			services.NewMaps,
			services.NewCommander,

			jobs.NewAlerts,

			handlers.NewWebhook,
			handlers.NewUpdate,

			server.NewMux,
			server.NewServer,
		),

		fx.Invoke(
			migrate,
			startAlertsJob,
			server.RegisterWebhook,
			server.RegisterListeningWebhooks,
			server.RegisterServer,
			clients.RegisterTelegram,
			clients.RegisterTelegramCommands,
		),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

func migrate(db clients.DB) error {
	err := db.AutoMigrate(
		&types2.Alert{},
		&types2.Notification{},
		&types2.Chat{},
		&types2.Map{},
	)
	if err != nil {
		return fmt.Errorf("db auto migrate trend: %w", err)
	}

	return nil
}

func startAlertsJob(lc fx.Lifecycle, alerts jobs.Alerts) {
	cctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := alerts.Run(cctx); err != nil {
				return fmt.Errorf("run: %w", err)
			}

			return nil
		},

		OnStop: func(context.Context) error {
			cancel()
			<-alerts.Done()

			return nil
		},
	})
}
