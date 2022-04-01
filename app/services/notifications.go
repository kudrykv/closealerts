package services

import (
	"closealerts/app/clients"
	"closealerts/app/repositories"
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type Notification struct {
	notification repositories.Notification
	log          *zap.SugaredLogger
	telegram     clients.Telegram
}

func NewNotification(
	log *zap.SugaredLogger,
	telegram clients.Telegram,
	notification repositories.Notification,
) Notification {
	return Notification{
		log:          log,
		telegram:     telegram,
		notification: notification,
	}
}

func (r Notification) Track(ctx context.Context, chatID int64, area string) error {
	if err := r.notification.Track(ctx, chatID, area); err != nil {
		return fmt.Errorf("track: %w", err)
	}

	return nil
}

func (r Notification) Tracking(ctx context.Context, id int64) (types2.Notifications, error) {
	list, err := r.notification.Tracking(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tracking: %w", err)
	}

	return list, nil
}

func (r Notification) Stop(ctx context.Context, id int64, area string) error {
	if err := r.notification.Stop(ctx, id, area); err != nil {
		return fmt.Errorf("stop: %w", err)
	}

	return nil
}

func (r Notification) Notify(ctx context.Context, alerts []types2.Alert) error {
	// 1. select ones who needs to be notified
	// 2. notify those
	// 3. mark notified
	// 4. unmark ones who doesn't match alerts

	eligible, err := r.notification.Eligible(ctx, alerts)
	if err != nil {
		return fmt.Errorf("eligible: %w", err)
	}

	alertsWg := r.notifyAboutAlertsAsync(ctx, eligible)

	endedFor, err := r.notification.AlertEnded(ctx, alerts)
	if err != nil {
		return fmt.Errorf("alert ended: %w", err)
	}

	endedAlertsWg := r.notifyAboutEndedAlertsAsync(ctx, endedFor)

	if err := r.notification.Unmark(ctx, alerts); err != nil {
		return fmt.Errorf("unmark: %w", err)
	}

	alertsWg.Wait()
	endedAlertsWg.Wait()

	return nil
}

func (r Notification) notifyAboutAlertsAsync(ctx context.Context, eligible types2.Notifications) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	go func() {
		sf := make(chan struct{}, 10)

		for _, notifications := range eligible.GroupByChatID() {
			sf <- struct{}{}
			wg.Add(1)

			go func(notifications types2.Notifications) {
				defer func() {
					<-sf
					wg.Done()
				}()

				if len(notifications) == 0 {
					return
				}

				if notifications[0].ChatID < 0 {
					return
				}

				r.telegram.MaybeSendText(ctx, notifications[0].ChatID, strings.Join(notifications.Areas(), ", ")+": тривога!")

				for _, notification := range notifications {
					if err := r.notification.Notified(ctx, notification); err != nil {
						r.log.Errorw("notified", "err", err)
					}
				}
			}(notifications)
		}
	}()

	return wg
}

func (r Notification) notifyAboutEndedAlertsAsync(ctx context.Context, endedFor types2.Notifications) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	if len(endedFor) == 0 {
		return wg
	}

	go func() {
		sf := make(chan struct{}, 10)
		for _, notifications := range endedFor.GroupByChatID() {
			sf <- struct{}{}
			wg.Add(1)

			go func(notifications types2.Notifications) {
				defer func() {
					<-sf
					wg.Done()
				}()

				if len(notifications) == 0 {
					return
				}

				if notifications[0].ChatID < 0 {
					return
				}

				r.telegram.MaybeSendText(ctx, notifications[0].ChatID, "тривога минула: "+strings.Join(notifications.Areas(), ", "))
			}(notifications)
		}
	}()

	return wg
}
