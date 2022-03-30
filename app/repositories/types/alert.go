package types

type Alert struct {
	ID   string `gorm:"column:id"`
	Type string `gorm:"column:type"`
}
