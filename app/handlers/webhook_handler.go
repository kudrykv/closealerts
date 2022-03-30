package handlers

import (
	"closealerts/app/types"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type WebhookHandler struct {
	Updates chan types.Update
}

func NewWebhook(config types.Config) WebhookHandler {
	return WebhookHandler{
		Updates: config.Updates,
	}
}

func (h WebhookHandler) Pipe(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	bts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(fmt.Errorf("read all: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var update types.Update
	if err = json.Unmarshal(bts, &update); err != nil {
		log.Println(fmt.Errorf("unmarshal: %w", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Updates <- update

	w.WriteHeader(http.StatusNoContent)
}
