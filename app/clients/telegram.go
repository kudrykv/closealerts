package clients

import (
	"closealerts/app/types"
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

type Telegram struct {
	log *zap.SugaredLogger

	Client *tgbotapi.BotAPI
	rl     ratelimit.Limiter
}

func NewTelegram(log *zap.SugaredLogger, config types.Config) (Telegram, error) {
	api, err := tgbotapi.NewBotAPI(config.TelegramBotAPI)
	if err != nil {
		return Telegram{}, fmt.Errorf("new bot api: %w", err)
	}

	api.Debug = config.DebugTelegram

	return Telegram{
		log:    log,
		Client: api,
		rl:     ratelimit.New(30, ratelimit.Per(time.Second)),
	}, nil
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

func RegisterTelegramCommands(log *zap.SugaredLogger, bot Telegram) error {
	commands := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "start",
			Description: "Коротко про те, як працює бот.",
		},

		tgbotapi.BotCommand{
			Command:     "areas",
			Description: "Список відслідковуваних областей, разом з налаштуванням",
		},

		tgbotapi.BotCommand{
			Command:     "alerts",
			Description: "Місця, де оголошена тривога",
		},

		tgbotapi.BotCommand{
			Command:     "map",
			Description: "Показати тривоги на мапі",
		},

		tgbotapi.BotCommand{
			Command:     "tracking",
			Description: "Список відслідковуваних місць",
		},
	)

	resp, err := bot.Client.Request(commands)
	if err != nil {
		log.Errorw("set commands", "err", err)
	} else {
		log.Infow(
			"set commands",
			"ok", resp.Ok,
			"desc", resp.Description,
		)
	}

	return nil
}

func (r Telegram) SetupWebhookEndpoint(pattern string, cert string) error {
	var (
		wh  tgbotapi.WebhookConfig
		err error
	)

	if len(cert) > 0 {
		wh, err = tgbotapi.NewWebhookWithCert(pattern, tgbotapi.FilePath(cert))
	} else {
		wh, err = tgbotapi.NewWebhook(pattern)
	}

	if err != nil {
		return fmt.Errorf("new webhook: %w", err)
	}

	if _, err = r.Client.Request(wh); err != nil {
		return fmt.Errorf("request: %w", err)
	}

	return nil
}

func (r Telegram) MaybeSend(_ context.Context, c tgbotapi.Chattable) {
	r.rl.Take()

	if _, err := r.Client.Request(c); err != nil {
		r.log.Errorw("send new message", "err", err)
	}
}

func (r Telegram) MaybeSendText(_ context.Context, chatID int64, msg string) {
	r.rl.Take()

	if _, err := r.Client.Send(tgbotapi.NewMessage(chatID, msg)); err != nil {
		r.log.Errorw("send new message", "err", err)
	}
}

func (r Telegram) Send(_ context.Context, chattable tgbotapi.Chattable) (tgbotapi.Message, error) {
	r.rl.Take()

	msg, err := r.Client.Send(chattable)
	if err != nil {
		return msg, fmt.Errorf("send: %w", err)
	}

	return msg, nil
}
