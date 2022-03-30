package clients

import (
	"closealerts/app/types"
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Telegram struct {
	log *zap.SugaredLogger

	Client *tgbotapi.BotAPI
}

func NewTelegram(log *zap.SugaredLogger, config types.Config) (Telegram, error) {
	api, err := tgbotapi.NewBotAPI(config.TelegramBotAPI)
	if err != nil {
		return Telegram{}, fmt.Errorf("new bot api: %w", err)
	}

	api.Debug = true

	return Telegram{log: log, Client: api}, nil
}

func RegisterTelegram(lc fx.Lifecycle, config types.Config, bot Telegram) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := bot.SetupWebhookEndpoint(config.WHEndpoint, config.Cert); err != nil {
				return fmt.Errorf("setup webhook endpoint: %w", err)
			}

			return nil
		},
	})
}

func (r Telegram) SetupWebhookEndpoint(pattern string, cert string) error {
	wh, err := tgbotapi.NewWebhookWithCert(pattern, tgbotapi.FilePath(cert))
	if err != nil {
		return fmt.Errorf("new webhook: %w", err)
	}

	if _, err = r.Client.Request(wh); err != nil {
		return fmt.Errorf("request: %w", err)
	}

	return nil
}

func (r Telegram) MaybeSendText(_ context.Context, chatID int64, msg string) {
	if _, err := r.Client.Send(tgbotapi.NewMessage(chatID, msg)); err != nil {
		r.log.Errorw("send new message", "err", err)
	}
}
