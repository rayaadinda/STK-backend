package menu

import "time"

type Menu struct {
	ID        string  `gorm:"type:uuid;primaryKey"`
	Name      string  `gorm:"type:varchar(120);not null"`
	ModuleKey string  `gorm:"column:module_key;type:varchar(120);not null;default:'systems/menus';index:idx_menus_scope_parent_order,priority:1"`
	ParentID  *string `gorm:"type:uuid;index:idx_menus_scope_parent_order,priority:2"`
	SortOrder int     `gorm:"not null;default:0;index:idx_menus_scope_parent_order,priority:3"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Menu) TableName() string {
	return "menus"
}
