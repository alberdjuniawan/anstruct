package core

import "context"

// Generator: natural language → Tree
// Dipakai oleh aistruct (prompt → blueprint Tree)
type Generator interface {
	FromPrompt(ctx context.Context, natural string) (*Tree, error)
}

// Parser: blueprint file ↔ Tree
// Dipakai oleh mstruct (struct file → folder) dan juga reverse/write
type Parser interface {
	Parse(ctx context.Context, blueprintPath string) (*Tree, error)
	Write(ctx context.Context, tree *Tree, path string) error
}

// Reverser: folder → Tree
// Dipakai oleh rstruct (folder → blueprint)
type Reverser interface {
	Reverse(ctx context.Context, inputDir string) (*Tree, error)
}

// Validator: Tree → error
// Validasi blueprint sebelum dieksekusi
type Validator interface {
	Validate(ctx context.Context, tree *Tree) error
}

// History: record & undo
// Semua operasi (create/reverse) dicatat untuk bisa di-undo
type History interface {
	Record(ctx context.Context, op Operation) error
	Undo(ctx context.Context) error
}

// --- Optional tambahan kontrak ---

// Watcher: sinkronisasi folder <-> blueprint
// (opsional, karena implementasi langsung di watcher package)
type Watcher interface {
	Run(ctx context.Context, cfg interface{},
		onFolder func(), onBlueprint func()) error
}
