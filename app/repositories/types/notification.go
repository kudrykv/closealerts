package types

type Notification struct {
	ChatID   int64  `gorm:"column:chat_id"`
	Area     string `gorm:"column:area"`
	Notified bool   `gorm:"column:notified"`
}
