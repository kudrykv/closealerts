package types

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Addr           string
	TickInterval   time.Duration
	SQLite3DBPath  string
	TelegramBotAPI string
	WHEndpoint     string
	Updates        chan Update
	Cert           string
	Key            string
	DebugTelegram  bool
}

func NewConfig() (Config, error) {
	tick, err := time.ParseDuration(os.Getenv("TICK_INTERVAL"))
	if err != nil {
		return Config{}, fmt.Errorf("parse duration: %w", err)
	}

	addr := "localhost:8080"
	if tmp := os.Getenv("SERVER_ADDR"); len(tmp) > 0 {
		addr = tmp
	}

	debugTelegram := strings.ToLower(os.Getenv("DEBUG_TELEGRAM")) == "true"

	return Config{
		SQLite3DBPath:  os.Getenv("SQLITE3_DB_PATH"),
		TickInterval:   tick,
		Addr:           addr,
		TelegramBotAPI: os.Getenv("TELEGRAM_BOT_API"),
		WHEndpoint:     os.Getenv("WEBHOOK_ENDPOINT"),
		Updates:        make(chan Update, 1000),
		Cert:           os.Getenv("SERVER_CERT"),
		Key:            os.Getenv("SERVER_KEY"),
		DebugTelegram:  debugTelegram,
	}, nil
}
