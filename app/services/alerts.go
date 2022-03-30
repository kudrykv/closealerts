package services

import (
	"closealerts/app/repositories"
	types2 "closealerts/app/repositories/types"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type Alerts struct {
	log    *zap.SugaredLogger
	alerts repositories.Alerts
}

func NewAlerts(log *zap.SugaredLogger, alerts repositories.Alerts) Alerts {
	return Alerts{log: log, alerts: alerts}
}

func (r Alerts) GetActive(ctx context.Context) ([]types2.Alert, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, "https://war-api.ukrzen.in.ua/alerts/api/v2/alerts/active.json", nil,
	)

	if err != nil {
		return nil, fmt.Errorf("new request with context: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io read all: %w", err)
	}

	var out AlertsResponse
	if err := json.Unmarshal(bts, &out); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	if len(out.Alerts) == 0 {
		return nil, nil
	}

	list := make([]types2.Alert, 0, len(out.Alerts))

	for _, alert := range out.Alerts {
		list = append(list, types2.Alert{ID: alert.Area, Type: alert.Type})
	}

	return list, nil
}

func (r Alerts) ReplaceAlerts(ctx context.Context, alerts []types2.Alert) error {
	if err := r.alerts.ReplaceAlerts(ctx, alerts); err != nil {
		return fmt.Errorf("replace alerts: %w", err)
	}

	return nil
}

type Alert struct {
	Type string `json:"t"`
	Area string `json:"n"`
}

type AlertsResponse struct {
	Alerts []Alert
}
