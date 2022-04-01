package jobs

import (
	"closealerts/app/services"
	"closealerts/app/types"
	"context"
	"time"

	"go.uber.org/zap"
)

type Alerts struct {
	tick         time.Duration
	done         chan struct{}
	alertSvc     services.Alerts
	log          *zap.SugaredLogger
	notification services.Notification
	fake         services.Fakes
}

func NewAlerts(
	log *zap.SugaredLogger,
	cfg types.Config,
	alertSvc services.Alerts,
	notification services.Notification,
	fake services.Fakes,
) Alerts {
	return Alerts{
		tick: cfg.TickInterval,
		done: make(chan struct{}),
		log:  log,

		fake:         fake,
		alertSvc:     alertSvc,
		notification: notification,
	}
}

func (r Alerts) Run(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(r.tick)
		defer func() { ticker.Stop() }()

		for {
			//ctx, span := otel.Tracer("job.alerts").Start(ctx, "run")

			select {
			case <-ctx.Done():
				close(r.done)
				return

			case <-ticker.C:
				alerts, err := r.alertSvc.GetActiveFromRemote(ctx)
				if err != nil {
					r.log.Errorw("get active alerts", "err", err)

					break
				}

				if alert, ok := r.fake.Alert(ctx); ok {
					alerts = append(alerts, alert)
				}

				if err := r.alertSvc.ReplaceAlerts(ctx, alerts); err != nil {
					r.log.Errorw("replace alerts", "err", err)

					break
				}

				if err := r.notification.Notify(ctx, alerts); err != nil {
					r.log.Errorw("notify", "err", err)

					break
				}
			}

			//span.End()
		}
	}()

	return nil
}

func (r Alerts) Done() <-chan struct{} {
	return r.done
}
