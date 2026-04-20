package database

import (
	"fmt"

	"gorm.io/gorm"

	"stk-backend/internal/menu"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&menu.Menu{}); err != nil {
		return fmt.Errorf("auto migrate menus table: %w", err)
	}

	const constraintsSQL = `
	DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM pg_constraint WHERE conname = 'menus_parent_fk'
		) THEN
			ALTER TABLE menus
				ADD CONSTRAINT menus_parent_fk
				FOREIGN KEY (parent_id)
				REFERENCES menus(id)
				ON DELETE CASCADE;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM pg_constraint WHERE conname = 'menus_name_not_blank'
		) THEN
			ALTER TABLE menus
				ADD CONSTRAINT menus_name_not_blank
				CHECK (char_length(trim(name)) > 0);
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM pg_constraint WHERE conname = 'menus_sort_order_non_negative'
		) THEN
			ALTER TABLE menus
				ADD CONSTRAINT menus_sort_order_non_negative
				CHECK (sort_order >= 0);
		END IF;
	END $$;
	`

	if err := db.Exec(constraintsSQL).Error; err != nil {
		return fmt.Errorf("apply menus constraints: %w", err)
	}

	const indexesSQL = `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_menus_root_sort_unique
		ON menus (sort_order)
		WHERE parent_id IS NULL;

	CREATE UNIQUE INDEX IF NOT EXISTS idx_menus_child_sort_unique
		ON menus (parent_id, sort_order)
		WHERE parent_id IS NOT NULL;
	`

	if err := db.Exec(indexesSQL).Error; err != nil {
		return fmt.Errorf("apply menus indexes: %w", err)
	}

	return nil
}
