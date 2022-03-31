package services

import (
	"closealerts/app/repositories"
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"
)

type Chats struct {
	chat repositories.Chats
}

func NewChats(chats repositories.Chats) Chats {
	return Chats{chat: chats}
}

func (r Chats) FirstOrCreate(ctx context.Context, id int64) (types2.Chat, error) {
	chat, err := r.chat.CreateOrSelect(ctx, id)
	if err != nil {
		return types2.Chat{}, fmt.Errorf("create or select: %w", err)
	}

	return chat, nil
}

func (r Chats) ClearCommand(ctx context.Context, id int64) error {
	if err := r.chat.ClearCommand(ctx, id); err != nil {
		return fmt.Errorf("clear command: %w", err)
	}

	return nil
}

func (r Chats) SetCommand(ctx context.Context, chatID int64, command string) error {
	if err := r.chat.SetCommand(ctx, chatID, command); err != nil {
		return fmt.Errorf("set command: %w", err)
	}

	return nil
}
