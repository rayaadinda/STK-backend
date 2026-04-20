package menu

type CreateInput struct {
	Scope    string
	Name     string
	ParentID *string
}

type UpdateInput struct {
	Scope string
	Name  string
}

type MoveInput struct {
	Scope    string
	ParentID *string
}

type ReorderInput struct {
	Scope    string
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
