package types

type Alert struct {
	ID   string `gorm:"column:id"`
	Type string `gorm:"column:type"`
}

type Alerts []Alert

func (r Alerts) Areas() []string {
	if len(r) == 0 {
		return nil
	}

	areas := make([]string, 0, len(r))
	for _, alert := range r {
		areas = append(areas, alert.ID)
	}

	return areas
}
