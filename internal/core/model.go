package core

type NodeType string

const (
	NodeDir  NodeType = "dir"
	NodeFile NodeType = "file"
)

type Node struct {
	Type         NodeType
	Name         string
	Content      string
	Children     []*Node
	OriginalName string
}

type Tree struct {
	Root *Node
}

type GenerateOptions struct {
	DryRun bool
	Force  bool
}

type AIOptions struct {
	Apply   bool
	DryRun  bool
	Verbose bool
	Retries int
	Force   bool
}

type Receipt struct {
	CreatedFiles []string
	CreatedDirs  []string
}

type OperationType string

const (
	OpCreate  OperationType = "create"
	OpReverse OperationType = "reverse"
	OpAI      OperationType = "ai_generate"
	OpAIApply OperationType = "ai_generate_apply"
)

type Operation struct {
	Type          OperationType
	Target        string
	Receipt       Receipt
	Timestamp     string
	BlueprintPath string // ADDED: Path to blueprint for recreation
	SourcePrompt  string // ADDED: Original AI prompt for recreation
}
