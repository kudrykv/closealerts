package server

import (
	"closealerts/app/handlers"
	"net/http"
)

func NewMux() *http.ServeMux {
	return http.NewServeMux()
}

func RegisterWebhook(mux *http.ServeMux, webhook handlers.WebhookHandler) {
	mux.HandleFunc("/tgwebhook", webhook.Pipe)
	mux.HandleFunc("/helloworld", webhook.HelloWorld)
}
