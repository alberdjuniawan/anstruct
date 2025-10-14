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
	ValidateWithOptions(ctx context.Context, tree *Tree, allowReserved bool) error
}

type History interface {
	Record(ctx context.Context, op Operation) error
	Undo(ctx context.Context) error
	Redo(ctx context.Context) error
	List(ctx context.Context) ([]Operation, error)
	Clear(ctx context.Context) error
}

type Watcher interface {
	Run(ctx context.Context, cfg interface{},
		onFolder func(), onBlueprint func()) error
}

type Service interface {
	AIStruct(ctx context.Context, prompt, outPath string, opts AIOptions) error
	MStruct(ctx context.Context, structFile, outputDir string, opts GenerateOptions) (Receipt, error)
	RStruct(ctx context.Context, inputDir string, outPath string) error
	NormalizeStruct(ctx context.Context, inputContent, outPath string, opts AIOptions) error
}
