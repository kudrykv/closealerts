package repositories

import (
	"closealerts/app/clients"
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"
)

type Chats struct {
	db clients.DB
}

func NewChats(db clients.DB) Chats {
	return Chats{db: db}
}

func (r Chats) CreateOrSelect(ctx context.Context, chat types2.Chat) (types2.Chat, error) {
	if err := r.db.DB().WithContext(ctx).FirstOrCreate(&chat).Error; err != nil {
		return types2.Chat{}, fmt.Errorf("first or create: %w", err)
	}

	return chat, nil
}

func (r Chats) ClearCommand(ctx context.Context, id int64) error {
	err := r.db.DB().WithContext(ctx).Model(&types2.Chat{}).Where("id = ?", id).UpdateColumn("command", "").Error
	if err != nil {
		return fmt.Errorf("clear command: %w", err)
	}

	return nil
}

func (r Chats) SetCommand(ctx context.Context, id int64, command string) error {
	err := r.db.DB().WithContext(ctx).Model(&types2.Chat{}).Where("id = ?", id).UpdateColumn("command", command).Error
	if err != nil {
		return fmt.Errorf("set command: %w", err)
	}

	return nil
}

func (r Chats) Grant(ctx context.Context, id int64, priv string) error {
	var col string

	switch priv {
	case "send_fake_event":
		col = "priv_send_fake_event"
	case "send_broadcast":
		col = "priv_broadcast"
	default:
		// silent noop
		return nil
	}

	err := r.db.DB().WithContext(ctx).Model(&types2.Chat{}).Where("id = ?", id).UpdateColumn(col, true).Error
	if err != nil {
		return fmt.Errorf("set priv %s: %w", priv, err)
	}

	return nil
}
