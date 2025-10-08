package core

// NodeType: dir/file
type NodeType string

const (
	NodeDir  NodeType = "dir"
	NodeFile NodeType = "file"
)

// Node: tree node
type Node struct {
	Type     NodeType
	Name     string
	Content  string
	Children []*Node
}

// Tree: root node
type Tree struct {
	Root *Node
}

// GenerateOptions: control file generation
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
