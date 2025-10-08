package core

import "context"

// Generator: natural language → Tree
type Generator interface {
	FromPrompt(ctx context.Context, natural string) (*Tree, error)
}

// Parser: blueprint file ↔ Tree
type Parser interface {
	Parse(ctx context.Context, blueprintPath string) (*Tree, error)
	Write(ctx context.Context, tree *Tree, path string) error
}

// Reverser: folder → Tree
type Reverser interface {
	Reverse(ctx context.Context, inputDir string) (*Tree, error)
}

// Validator: Tree → error
type Validator interface {
	Validate(ctx context.Context, tree *Tree) error
}

// History: record & undo
type History interface {
	Record(ctx context.Context, op Operation) error
	Undo(ctx context.Context) error
}
