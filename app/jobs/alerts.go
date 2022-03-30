package jobs

import (
	"closealerts/app/services"
	"closealerts/app/types"
	"context"
	"time"

	"go.uber.org/zap"
)

type Alerts struct {
	tick     time.Duration
	done     chan struct{}
	alertSvc services.Alerts
	log      *zap.SugaredLogger
}

func NewAlerts(log *zap.SugaredLogger, cfg types.Config, alertSvc services.Alerts) Alerts {
	return Alerts{
		tick: cfg.TickInterval,
		done: make(chan struct{}),
		log:  log,

		alertSvc: alertSvc,
	}
}

func (r Alerts) Run(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(r.tick)
		defer func() { ticker.Stop() }()

		for {
			select {
			case <-ctx.Done():
				close(r.done)
				return

			case <-ticker.C:
				alerts, err := r.alertSvc.GetActive(ctx)
				if err != nil {
					r.log.Errorw("get active alerts", "err", err)

					break
				}

				if err := r.alertSvc.ReplaceAlerts(ctx, alerts); err != nil {
					r.log.Errorw("replace alerts", "err", err)
				}
			}
		}
	}()

	return nil
}

func (r Alerts) Done() <-chan struct{} {
	return r.done
}
