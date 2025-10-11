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

// AIOptions: opsi untuk AI generation
type AIOptions struct {
	Apply   bool // langsung generate folder dari hasil AI
	DryRun  bool // preview saja tanpa tulis file
	Verbose bool // tampilkan output mentah AI
	Retries int  // jumlah retry jika output invalid
	Force   bool // overwrite file yang sudah ada
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
	OpAI      OperationType = "ai_generate"       // aistruct generate
	OpAIApply OperationType = "ai_generate_apply" // aistruct --apply
)

// Operation: dicatat di history
type Operation struct {
	Type      OperationType
	Target    string  // path utama (file atau folder)
	Receipt   Receipt // detail yang dibuat (untuk undo)
	Timestamp string  // waktu operasi
}
