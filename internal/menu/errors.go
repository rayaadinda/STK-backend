package menu

import "errors"

var (
	ErrMenuNotFound       = errors.New("menu not found")
	ErrMenuNameInvalid    = errors.New("menu name is invalid")
	ErrParentMenuNotFound = errors.New("parent menu not found")
	ErrInvalidMoveTarget  = errors.New("invalid move target")
	ErrInvalidPosition    = errors.New("invalid reorder position")
)
