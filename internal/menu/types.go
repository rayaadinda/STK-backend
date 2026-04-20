package menu

type CreateInput struct {
	Name     string
	ParentID *string
}

type UpdateInput struct {
	Name string
}

type MoveInput struct {
	ParentID *string
}

type ReorderInput struct {
	ParentID *string
	Position int
}

type Item struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID *string `json:"parentId,omitempty"`
	Position int     `json:"position"`
}

type TreeNode struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	ParentID *string     `json:"parentId,omitempty"`
	Position int         `json:"position"`
	Children []*TreeNode `json:"children,omitempty"`
}
