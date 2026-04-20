package menu

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetTree(ctx context.Context) ([]*TreeNode, error) {
	menus, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	roots, _ := buildTree(menus)
	return roots, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*TreeNode, error) {
	menus, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	_, nodeMap := buildTree(menus)
	node, ok := nodeMap[id]
	if !ok {
		return nil, ErrMenuNotFound
	}

	return node, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Item, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Item{}, ErrMenuNameInvalid
	}

	var created Menu
	err := s.repo.Transaction(ctx, func(tx *Repository) error {
		if input.ParentID != nil {
			exists, err := tx.ExistsByID(ctx, *input.ParentID)
			if err != nil {
				return err
			}
			if !exists {
				return ErrParentMenuNotFound
			}
		}

		position, err := tx.CountByParent(ctx, input.ParentID)
		if err != nil {
			return err
		}

		created = Menu{
			ID:        uuid.NewString(),
			Name:      name,
			ParentID:  input.ParentID,
			SortOrder: position,
		}

		if err := tx.Create(ctx, &created); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return Item{}, err
	}

	return toItem(created), nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Item, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Item{}, ErrMenuNameInvalid
	}

	var updated Menu
	err := s.repo.Transaction(ctx, func(tx *Repository) error {
		existing, err := tx.GetByID(ctx, id)
		if err != nil {
			return err
		}

		if existing.Name == name {
			updated = existing
			return nil
		}

		if err := tx.UpdateName(ctx, id, name); err != nil {
			return err
		}

		updated, err = tx.GetByID(ctx, id)
		return err
	})
	if err != nil {
		return Item{}, err
	}

	return toItem(updated), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Transaction(ctx, func(tx *Repository) error {
		target, err := tx.GetByID(ctx, id)
		if err != nil {
			return err
		}

		if err := tx.DeleteByID(ctx, id); err != nil {
			return err
		}

		if err := tx.ShiftOrdersFrom(ctx, target.ParentID, target.SortOrder+1, -1); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) Move(ctx context.Context, id string, input MoveInput) (Item, error) {
	var moved Menu
	err := s.repo.Transaction(ctx, func(tx *Repository) error {
		target, err := tx.GetByID(ctx, id)
		if err != nil {
			return err
		}

		if sameParent(target.ParentID, input.ParentID) {
			moved = target
			return nil
		}

		if input.ParentID != nil {
			if *input.ParentID == id {
				return ErrInvalidMoveTarget
			}

			exists, err := tx.ExistsByID(ctx, *input.ParentID)
			if err != nil {
				return err
			}
			if !exists {
				return ErrParentMenuNotFound
			}

			isDescendant, err := tx.IsDescendant(ctx, id, *input.ParentID)
			if err != nil {
				return err
			}
			if isDescendant {
				return ErrInvalidMoveTarget
			}
		}

		parkingOrder, err := tx.MaxOrderByParent(ctx, target.ParentID)
		if err != nil {
			return err
		}

		if err := tx.UpdateOrder(ctx, id, parkingOrder+1); err != nil {
			return err
		}

		if err := tx.ShiftOrdersFrom(ctx, target.ParentID, target.SortOrder+1, -1); err != nil {
			return err
		}

		nextPosition, err := tx.CountByParent(ctx, input.ParentID)
		if err != nil {
			return err
		}

		if err := tx.UpdateParentAndOrder(ctx, id, input.ParentID, nextPosition); err != nil {
			return err
		}

		moved, err = tx.GetByID(ctx, id)
		return err
	})
	if err != nil {
		return Item{}, err
	}

	return toItem(moved), nil
}

func (s *Service) Reorder(ctx context.Context, id string, input ReorderInput) (Item, error) {
	if input.Position < 0 {
		return Item{}, ErrInvalidPosition
	}

	var reordered Menu
	err := s.repo.Transaction(ctx, func(tx *Repository) error {
		target, err := tx.GetByID(ctx, id)
		if err != nil {
			return err
		}

		if input.ParentID != nil {
			if *input.ParentID == id {
				return ErrInvalidMoveTarget
			}

			exists, err := tx.ExistsByID(ctx, *input.ParentID)
			if err != nil {
				return err
			}
			if !exists {
				return ErrParentMenuNotFound
			}

			isDescendant, err := tx.IsDescendant(ctx, id, *input.ParentID)
			if err != nil {
				return err
			}
			if isDescendant {
				return ErrInvalidMoveTarget
			}
		}

		sameLevel := sameParent(target.ParentID, input.ParentID)
		if sameLevel {
			siblingCount, err := tx.CountByParent(ctx, input.ParentID)
			if err != nil {
				return err
			}

			if siblingCount == 0 {
				reordered = target
				return nil
			}

			newPosition := clamp(input.Position, 0, siblingCount-1)
			if newPosition == target.SortOrder {
				reordered = target
				return nil
			}

			parkingOrder, err := tx.MaxOrderByParent(ctx, target.ParentID)
			if err != nil {
				return err
			}

			if err := tx.UpdateOrder(ctx, id, parkingOrder+1); err != nil {
				return err
			}

			if newPosition < target.SortOrder {
				if err := tx.ShiftOrdersRange(ctx, input.ParentID, newPosition, target.SortOrder-1, 1); err != nil {
					return err
				}
			} else {
				if err := tx.ShiftOrdersRange(ctx, input.ParentID, target.SortOrder+1, newPosition, -1); err != nil {
					return err
				}
			}

			if err := tx.UpdateOrder(ctx, id, newPosition); err != nil {
				return err
			}
		} else {
			parkingOrder, err := tx.MaxOrderByParent(ctx, target.ParentID)
			if err != nil {
				return err
			}

			if err := tx.UpdateOrder(ctx, id, parkingOrder+1); err != nil {
				return err
			}

			if err := tx.ShiftOrdersFrom(ctx, target.ParentID, target.SortOrder+1, -1); err != nil {
				return err
			}

			siblingCount, err := tx.CountByParent(ctx, input.ParentID)
			if err != nil {
				return err
			}

			newPosition := clamp(input.Position, 0, siblingCount)
			if err := tx.ShiftOrdersFrom(ctx, input.ParentID, newPosition, 1); err != nil {
				return err
			}

			if err := tx.UpdateParentAndOrder(ctx, id, input.ParentID, newPosition); err != nil {
				return err
			}
		}

		reordered, err = tx.GetByID(ctx, id)
		return err
	})
	if err != nil {
		return Item{}, err
	}

	return toItem(reordered), nil
}

func ValidateUUID(value string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}
	return nil
}

func toItem(entity Menu) Item {
	return Item{
		ID:       entity.ID,
		Name:     entity.Name,
		ParentID: entity.ParentID,
		Position: entity.SortOrder,
	}
}

func buildTree(entities []Menu) ([]*TreeNode, map[string]*TreeNode) {
	nodeMap := make(map[string]*TreeNode, len(entities))
	for _, entity := range entities {
		nodeMap[entity.ID] = &TreeNode{
			ID:       entity.ID,
			Name:     entity.Name,
			ParentID: entity.ParentID,
			Position: entity.SortOrder,
			Children: make([]*TreeNode, 0),
		}
	}

	roots := make([]*TreeNode, 0)
	for _, entity := range entities {
		node := nodeMap[entity.ID]
		if entity.ParentID == nil {
			roots = append(roots, node)
			continue
		}

		parentNode, ok := nodeMap[*entity.ParentID]
		if !ok {
			roots = append(roots, node)
			continue
		}

		parentNode.Children = append(parentNode.Children, node)
	}

	sortNodes(roots)
	return roots, nodeMap
}

func sortNodes(nodes []*TreeNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Position == nodes[j].Position {
			return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
		}
		return nodes[i].Position < nodes[j].Position
	})

	for _, node := range nodes {
		sortNodes(node.Children)
	}
}

func sameParent(currentParentID, nextParentID *string) bool {
	if currentParentID == nil && nextParentID == nil {
		return true
	}

	if currentParentID == nil || nextParentID == nil {
		return false
	}

	return *currentParentID == *nextParentID
}

func clamp(value, min, max int) int {
	if max < min {
		return min
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func IsKnownDomainError(err error) bool {
	return errors.Is(err, ErrMenuNotFound) ||
		errors.Is(err, ErrMenuNameInvalid) ||
		errors.Is(err, ErrParentMenuNotFound) ||
		errors.Is(err, ErrInvalidMoveTarget) ||
		errors.Is(err, ErrInvalidPosition)
}
