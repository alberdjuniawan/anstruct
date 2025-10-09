package core

// NodeType: dir/file
type NodeType string

const (
	NodeDir  NodeType = "dir"
	NodeFile NodeType = "file"
)

// Node: node dalam tree blueprint
type Node struct {
	Type         NodeType
	Name         string
	Content      string
	Children     []*Node
	OriginalName string
}

// Tree: representasi blueprint project
type Tree struct {
	Root *Node
}

// GenerateOptions: opsi saat generate folder
type GenerateOptions struct {
	DryRun bool
	Force  bool
}

// Receipt: hasil generate
type Receipt struct {
	CreatedFiles []string
	CreatedDirs  []string
}

// OperationType: jenis operasi
type OperationType string

const (
	OpCreate  OperationType = "create"
	OpReverse OperationType = "reverse"
)

// Operation: dicatat di history
type Operation struct {
	Type    OperationType
	Target  string
	Receipt Receipt
}
