-- Reference SQL migration for menus table.
-- Runtime migration is executed in internal/database/migrate.go.

CREATE TABLE IF NOT EXISTS menus (
  id UUID PRIMARY KEY,
  name VARCHAR(120) NOT NULL,
  parent_id UUID NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT menus_parent_fk
    FOREIGN KEY (parent_id)
    REFERENCES menus(id)
    ON DELETE CASCADE,
  CONSTRAINT menus_name_not_blank
    CHECK (char_length(trim(name)) > 0),
  CONSTRAINT menus_sort_order_non_negative
    CHECK (sort_order >= 0)
);

CREATE INDEX IF NOT EXISTS idx_menus_parent_order
  ON menus(parent_id, sort_order);

CREATE UNIQUE INDEX IF NOT EXISTS idx_menus_root_sort_unique
  ON menus(sort_order)
  WHERE parent_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_menus_child_sort_unique
  ON menus(parent_id, sort_order)
  WHERE parent_id IS NOT NULL;
