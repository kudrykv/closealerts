package services

import (
	"closealerts/app/clients"
	"closealerts/app/repositories"
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"
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

func (n Notification) Track(ctx context.Context, chatID int64, area string) error {
	if err := n.notification.Track(ctx, chatID, area); err != nil {
		return fmt.Errorf("track: %w", err)
	}

	return nil
}

func (n Notification) Tracking(ctx context.Context, id int64) ([]types2.Notification, error) {
	list, err := n.notification.Tracking(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tracking: %w", err)
	}

	return list, nil
}

func (n Notification) Stop(ctx context.Context, id int64, area string) error {
	if err := n.notification.Stop(ctx, id, area); err != nil {
		return fmt.Errorf("stop: %w", err)
	}

	return nil
}

func (n Notification) Notify(ctx context.Context, alerts []types2.Alert) error {
	// 1. select ones who needs to be notified
	// 2. notify those
	// 3. mark notified
	// 4. unmark ones who doesn't match alerts

	eligible, err := n.notification.Eligible(ctx, alerts)
	if err != nil {
		return fmt.Errorf("eligible: %w", err)
	}

	sf := make(chan struct{}, 4)
	wg := sync.WaitGroup{}

	for _, notification := range eligible {
		sf <- struct{}{}
		wg.Add(1)

		go func(notification types2.Notification) {
			defer func() {
				<-sf
				wg.Done()
			}()

			n.telegram.MaybeSendText(ctx, notification.ChatID, notification.Area+": тривога!")
		}(notification)
	}

	wg.Wait()

	if err := n.notification.Notified(ctx, eligible); err != nil {
		return fmt.Errorf("mark as notified: %w", err)
	}

	if err := n.notification.Unmark(ctx, alerts); err != nil {
		return fmt.Errorf("unmark: %w", err)
	}

	return nil
}
