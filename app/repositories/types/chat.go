package types

type Chat struct {
	ID       int64  `gorm:"column:id"`
	Username string `gorm:"column:username"`
	Command  string `gorm:"column:command"`

	PrivSendFakeEvent bool `gorm:"column:priv_send_fake_event"`
}
