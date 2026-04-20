package menu

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const reorderTempOffset = 1000000

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Transaction(ctx context.Context, fn func(txRepo *Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
}

func (r *Repository) List(ctx context.Context) ([]Menu, error) {
	var menus []Menu
	if err := r.db.WithContext(ctx).
		Order("CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END").
		Order("parent_id ASC").
		Order("sort_order ASC").
		Order("created_at ASC").
		Find(&menus).Error; err != nil {
		return nil, fmt.Errorf("list menus: %w", err)
	}

	return menus, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Menu, error) {
	var entity Menu
	if err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Menu{}, ErrMenuNotFound
		}
		return Menu{}, fmt.Errorf("get menu by id: %w", err)
	}

	return entity, nil
}

func (r *Repository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&Menu{}).
		Where("id = ?", id).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check menu existence: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) Create(ctx context.Context, entity *Menu) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("create menu: %w", err)
	}
	return nil
}

func (r *Repository) UpdateName(ctx context.Context, id, name string) error {
	result := r.db.WithContext(ctx).
		Model(&Menu{}).
		Where("id = ?", id).
		Update("name", name)
	if result.Error != nil {
		return fmt.Errorf("update menu name: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrMenuNotFound
	}
	return nil
}

func (r *Repository) DeleteByID(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Menu{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete menu: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrMenuNotFound
	}
	return nil
}

func (r *Repository) CountByParent(ctx context.Context, parentID *string) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&Menu{})
	query = scopeParent(query, parentID)

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count siblings: %w", err)
	}

	return int(count), nil
}

func (r *Repository) MaxOrderByParent(ctx context.Context, parentID *string) (int, error) {
	query := r.db.WithContext(ctx).Model(&Menu{}).Select("COALESCE(MAX(sort_order), -1)")
	query = scopeParent(query, parentID)

	var maxOrder int
	if err := query.Scan(&maxOrder).Error; err != nil {
		return 0, fmt.Errorf("max sort order by parent: %w", err)
	}

	return maxOrder, nil
}

func (r *Repository) ShiftOrdersFrom(ctx context.Context, parentID *string, from int, delta int) error {
	if delta == 0 {
		return nil
	}

	if err := r.shiftOrdersWithTemporaryOffset(ctx, parentID, "sort_order >= ?", []any{from}, delta); err != nil {
		return fmt.Errorf("shift orders from index: %w", err)
	}
	return nil
}

func (r *Repository) ShiftOrdersRange(ctx context.Context, parentID *string, start, end, delta int) error {
	if delta == 0 || start > end {
		return nil
	}

	if err := r.shiftOrdersWithTemporaryOffset(
		ctx,
		parentID,
		"sort_order >= ? AND sort_order <= ?",
		[]any{start, end},
		delta,
	); err != nil {
		return fmt.Errorf("shift orders in range: %w", err)
	}
	return nil
}

func (r *Repository) UpdateParentAndOrder(ctx context.Context, id string, parentID *string, order int) error {
	result := r.db.WithContext(ctx).
		Model(&Menu{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"parent_id":  parentID,
			"sort_order": order,
		})
	if result.Error != nil {
		return fmt.Errorf("update parent and order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrMenuNotFound
	}
	return nil
}

func (r *Repository) UpdateOrder(ctx context.Context, id string, order int) error {
	result := r.db.WithContext(ctx).
		Model(&Menu{}).
		Where("id = ?", id).
		Update("sort_order", order)
	if result.Error != nil {
		return fmt.Errorf("update order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrMenuNotFound
	}
	return nil
}

func (r *Repository) IsDescendant(ctx context.Context, ancestorID, targetID string) (bool, error) {
	const query = `
		WITH RECURSIVE descendants AS (
			SELECT id, parent_id
			FROM menus
			WHERE id = @ancestor
			UNION ALL
			SELECT m.id, m.parent_id
			FROM menus m
			INNER JOIN descendants d ON m.parent_id = d.id
		)
		SELECT EXISTS (
			SELECT 1
			FROM descendants
			WHERE id = @target
		)
	`

	var exists bool
	if err := r.db.WithContext(ctx).
		Raw(query, map[string]any{"ancestor": ancestorID, "target": targetID}).
		Scan(&exists).Error; err != nil {
		return false, fmt.Errorf("check descendant relationship: %w", err)
	}

	return exists, nil
}

func scopeParent(query *gorm.DB, parentID *string) *gorm.DB {
	if parentID == nil {
		return query.Where("parent_id IS NULL")
	}

	return query.Where("parent_id = ?", *parentID)
}

func (r *Repository) shiftOrdersWithTemporaryOffset(
	ctx context.Context,
	parentID *string,
	condition string,
	conditionArgs []any,
	delta int,
) error {
	baseQuery := r.db.WithContext(ctx).Model(&Menu{}).Where(condition, conditionArgs...)
	baseQuery = scopeParent(baseQuery, parentID)

	var ids []string
	if err := baseQuery.Pluck("id", &ids).Error; err != nil {
		return fmt.Errorf("load affected menu ids: %w", err)
	}

	if len(ids) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).
		Model(&Menu{}).
		Where("id IN ?", ids).
		Update("sort_order", gorm.Expr("sort_order + ?", reorderTempOffset)).Error; err != nil {
		return fmt.Errorf("apply temporary shift: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Model(&Menu{}).
		Where("id IN ?", ids).
		Update(
			"sort_order",
			gorm.Expr("sort_order - ? + ?", reorderTempOffset, delta),
		).Error; err != nil {
		return fmt.Errorf("normalize shifted values: %w", err)
	}

	return nil
}
