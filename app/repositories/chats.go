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

func (r Chats) CreateOrSelect(ctx context.Context, id int64) (types2.Chat, error) {
	chat := types2.Chat{ID: id}
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
