package types

import "sort"

type Notification struct {
	ChatID   int64  `gorm:"column:chat_id"`
	Area     string `gorm:"column:area"`
	Notified bool   `gorm:"column:notified"`
}

type Notifications []Notification

func (r Notifications) Areas() []string {
	if len(r) == 0 {
		return nil
	}

	areas := make([]string, 0, len(r))
	for _, notification := range r {
		areas = append(areas, notification.Area)
	}

	return areas
}

func (r Notifications) GroupByChatID() map[int64]Notifications {
	if len(r) == 0 {
		return nil
	}

	cp := make(Notifications, len(r))
	copy(cp, r)
	sort.Slice(cp, func(i, j int) bool { return cp[i].ChatID < cp[j].ChatID })

	out := make(map[int64]Notifications, len(r)/4)
	ptrChatID := cp[0].ChatID
	start := 0
	end := 0

	for _, notification := range cp {
		if notification.ChatID == ptrChatID {
			end++

			continue
		}

		out[ptrChatID] = cp[start:end]
		start = end
		ptrChatID = notification.ChatID
	}

	out[ptrChatID] = cp[start:]

	return out
}
