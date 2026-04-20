package menu

import "time"

type Menu struct {
	ID        string  `gorm:"type:uuid;primaryKey"`
	Name      string  `gorm:"type:varchar(120);not null"`
	ParentID  *string `gorm:"type:uuid;index:idx_menus_parent_order,priority:1"`
	SortOrder int     `gorm:"not null;default:0;index:idx_menus_parent_order,priority:2"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Menu) TableName() string {
	return "menus"
}
