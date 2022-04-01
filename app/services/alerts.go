package services

import (
	"closealerts/app/repositories"
	types2 "closealerts/app/repositories/types"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type Alerts struct {
	log    *zap.SugaredLogger
	alerts repositories.Alerts
}

func NewAlerts(log *zap.SugaredLogger, alerts repositories.Alerts) Alerts {
	return Alerts{log: log, alerts: alerts}
}

func (r Alerts) GetActiveFromRemote(ctx context.Context) ([]types2.Alert, error) {
	list, err := r.ukrzen(ctx)
	if err != nil {
		r.log.Errorw("active alerts", "source", "ukrzen", "err", err)
	} else {
		r.log.Infow("active alerts", "source", "ukrzen")

		return list, nil
	}

	list, err = r.vadimklimenko(ctx)
	if err != nil {
		r.log.Errorw("active alerts", "source", "ukrzen", "err", err)
	} else {
		r.log.Infow("active alerts", "source", "vadimklimenko")

		return list, nil
	}

	list, err = r.alarmmap(ctx)
	if err != nil {
		return nil, fmt.Errorf("alarmmap: %w", err)
	}

	r.log.Infow("active alerts", "source", "alarmmap")

	return list, nil
}

func (r Alerts) ReplaceAlerts(ctx context.Context, alerts []types2.Alert) error {
	if err := r.alerts.ReplaceAlerts(ctx, alerts); err != nil {
		return fmt.Errorf("replace alerts: %w", err)
	}

	r.log.Debugw("replaced alerts with active ones")

	return nil
}

func (r Alerts) GetActive(ctx context.Context) (types2.Alerts, error) {
	list, err := r.alerts.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active: %w", err)
	}

	return list, nil
}

type Alert struct {
	Type string `json:"t"`
	Area string `json:"n"`
}

type AlertsResponse struct {
	Alerts []Alert
}

func (r Alerts) ukrzen(ctx context.Context) ([]types2.Alert, error) {
	var (
		resp AlertsResponse
		err  error
	)

	if err = mkReqUnmarshal(ctx, "https://war-api.ukrzen.in.ua/alerts/api/v2/alerts/active.json", &resp); err != nil {
		return nil, fmt.Errorf("mk req: %w", err)
	}

	if len(resp.Alerts) == 0 {
		r.log.Info("no active alerts")

		return nil, nil
	}

	list := make([]types2.Alert, 0, len(resp.Alerts))

	for _, alert := range resp.Alerts {
		list = append(list, types2.Alert{ID: alert.Area, Type: alert.Type})
	}

	r.log.Infow("active from remote", "list", list)

	return list, nil
}

type AlarmMapResponseItem struct {
	District string `json:"district"`
}

func (r Alerts) alarmmap(ctx context.Context) ([]types2.Alert, error) {
	var (
		resp []AlarmMapResponseItem
		err  error
	)

	if err = mkReqUnmarshal(ctx, "https://alarmmap.online/assets/alerts.json", &resp); err != nil {
		return nil, fmt.Errorf("mk req: %w", err)
	}

	if len(resp) == 0 {
		return nil, nil
	}

	list := make([]types2.Alert, 0, len(resp))

	for _, alert := range resp {
		area := strings.SplitN(alert.District, "_", 2)[0]
		if area == "ІваноФранківська" {
			area = "Івано-Франківська"
		}

		list = append(list, types2.Alert{ID: alert.District})
	}

	r.log.Infow("active from remote", "list", list)

	return list, nil
}

func mkReqUnmarshal(ctx context.Context, url string, dst interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return fmt.Errorf("new request with context: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io read all: %w", err)
	}
	_ = resp.Body.Close()

	if err := json.Unmarshal(bts, dst); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	return nil
}

type VadimResponse struct {
	States  map[string]VadimArea `json:"states"`
	Enabled bool                 `json:"enabled"`
}

type VadimArea struct {
	Enabled   bool                   `json:"enabled"`
	Type      string                 `json:"type:"`
	Districts map[string]VadimRegion `json:"districts"`
}

type VadimRegion struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
}

func (r Alerts) vadimklimenko(ctx context.Context) (types2.Alerts, error) {
	var (
		resp VadimResponse
		err  error
	)

	if err = mkReqUnmarshal(ctx, "https://emapa.fra1.cdn.digitaloceanspaces.com/statuses.json", &resp); err != nil {
		return nil, fmt.Errorf("mk req: %w", err)
	}

	var list types2.Alerts

	for state, data := range resp.States {
		area := strings.SplitN(state, " ", 2)[0]
		if state == "м. Київ" {
			area = "м. Київ"
		}

		if data.Enabled {
			list = append(list, types2.Alert{ID: area, Type: "o"})
		}

		for region, data := range data.Districts {
			area := strings.SplitN(region, " ", 2)[0]

			if data.Enabled {
				list = append(list, types2.Alert{ID: area, Type: "r"})
			}
		}
	}

	r.log.Infow("active from remote", "list", list)

	return list, nil
}
