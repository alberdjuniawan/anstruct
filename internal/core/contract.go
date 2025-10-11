package core

import "context"

type Generator interface {
	FromPrompt(ctx context.Context, natural string) (*Tree, error)
}

type Parser interface {
	Parse(ctx context.Context, blueprintPath string) (*Tree, error)
	Write(ctx context.Context, tree *Tree, path string) error
	ParseString(ctx context.Context, content string) (*Tree, error)
}

type Reverser interface {
	Reverse(ctx context.Context, inputDir string) (*Tree, error)
}

type Validator interface {
	Validate(ctx context.Context, tree *Tree) error
}

type History interface {
	Record(ctx context.Context, op Operation) error
	Undo(ctx context.Context) error
}

type Watcher interface {
	Run(ctx context.Context, cfg interface{},
		onFolder func(), onBlueprint func()) error
}
