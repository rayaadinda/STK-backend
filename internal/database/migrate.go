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

	const scopeSQL = `
	ALTER TABLE menus
		ADD COLUMN IF NOT EXISTS module_key VARCHAR(120);

	UPDATE menus
	SET module_key = 'systems/menus'
	WHERE module_key IS NULL OR char_length(trim(module_key)) = 0;

	ALTER TABLE menus
		ALTER COLUMN module_key SET DEFAULT 'systems/menus';

	ALTER TABLE menus
		ALTER COLUMN module_key SET NOT NULL;
	`

	if err := db.Exec(scopeSQL).Error; err != nil {
		return fmt.Errorf("apply menus scope column: %w", err)
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

		IF NOT EXISTS (
			SELECT 1 FROM pg_constraint WHERE conname = 'menus_module_key_not_blank'
		) THEN
			ALTER TABLE menus
				ADD CONSTRAINT menus_module_key_not_blank
				CHECK (char_length(trim(module_key)) > 0);
		END IF;
	END $$;
	`

	if err := db.Exec(constraintsSQL).Error; err != nil {
		return fmt.Errorf("apply menus constraints: %w", err)
	}

	const indexesSQL = `
	DROP INDEX IF EXISTS idx_menus_root_sort_unique;
	DROP INDEX IF EXISTS idx_menus_child_sort_unique;

	CREATE UNIQUE INDEX IF NOT EXISTS idx_menus_scope_root_sort_unique
		ON menus (module_key, sort_order)
		WHERE parent_id IS NULL;

	CREATE UNIQUE INDEX IF NOT EXISTS idx_menus_scope_child_sort_unique
		ON menus (module_key, parent_id, sort_order)
		WHERE parent_id IS NOT NULL;
	`

	if err := db.Exec(indexesSQL).Error; err != nil {
		return fmt.Errorf("apply menus indexes: %w", err)
	}

	return nil
}
