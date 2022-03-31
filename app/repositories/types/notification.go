package types

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
