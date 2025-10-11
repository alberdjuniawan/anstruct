package history

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type History struct {
	LogPath string
}

func New(logPath string) *History {
	return &History{LogPath: logPath}
}

// Record: simpan operasi ke log JSON
func (h *History) Record(ctx context.Context, op core.Operation) error {
	// Add timestamp
	if op.Timestamp == "" {
		op.Timestamp = time.Now().Format(time.RFC3339)
	}

	_ = os.MkdirAll(filepath.Dir(h.LogPath), 0o755)
	f, err := os.OpenFile(h.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(op)
}

// Undo: rollback operasi terakhir
func (h *History) Undo(ctx context.Context) error {
	data, err := os.ReadFile(h.LogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return core.ErrHistoryEmpty
		}
		return err
	}

	lines := splitLines(string(data))
	if len(lines) == 0 {
		return core.ErrHistoryEmpty
	}

	last := lines[len(lines)-1]
	var op core.Operation
	if err := json.Unmarshal([]byte(last), &op); err != nil {
		return err
	}

	// Execute undo based on operation type
	if err := h.undoOperation(op); err != nil {
		return err
	}

	// Remove last line from log
	return truncateLastLine(h.LogPath)
}

// undoOperation performs the actual undo logic
func (h *History) undoOperation(op core.Operation) error {
	switch op.Type {
	case core.OpCreate:
		// Undo mstruct: hapus folder yang di-generate
		return h.undoCreate(op)

	case core.OpReverse:
		// Undo rstruct: hapus .struct file
		return h.undoReverse(op)

	case core.OpAI:
		// Undo aistruct (blueprint mode): hapus .struct file
		return h.undoAIBlueprint(op)

	case core.OpAIApply:
		// Undo aistruct --apply: hapus folder yang di-generate
		return h.undoAIApply(op)

	default:
		return errors.New("unsupported undo type: " + string(op.Type))
	}
}

// undoCreate: rollback OpCreate (mstruct)
func (h *History) undoCreate(op core.Operation) error {
	// Hapus files dulu (bottom-up)
	for _, f := range op.Receipt.CreatedFiles {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			// Log error but continue
		}
	}

	// Hapus directories (dari dalam ke luar)
	dirs := op.Receipt.CreatedDirs
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	for _, d := range dirs {
		if err := os.Remove(d); err != nil && !os.IsNotExist(err) {
			// Directory might not be empty, that's ok
		}
	}

	return nil
}

// undoReverse: rollback OpReverse (rstruct)
func (h *History) undoReverse(op core.Operation) error {
	return os.Remove(op.Target)
}

// undoAIBlueprint: rollback OpAI (aistruct blueprint mode)
func (h *History) undoAIBlueprint(op core.Operation) error {
	// Simply delete the .struct file
	return os.Remove(op.Target)
}

// undoAIApply: rollback OpAIApply (aistruct --apply)
func (h *History) undoAIApply(op core.Operation) error {
	// Same as undoCreate - remove all generated files/folders
	return h.undoCreate(op)
}

// List: tampilkan history
func (h *History) List(ctx context.Context) ([]core.Operation, error) {
	data, err := os.ReadFile(h.LogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []core.Operation{}, nil
		}
		return nil, err
	}

	lines := splitLines(string(data))
	ops := make([]core.Operation, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		var op core.Operation
		if err := json.Unmarshal([]byte(line), &op); err != nil {
			continue // Skip invalid lines
		}
		ops = append(ops, op)
	}

	return ops, nil
}

// Clear: hapus semua history
func (h *History) Clear(ctx context.Context) error {
	return os.Remove(h.LogPath)
}

// --- Helpers ---

func truncateLastLine(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := splitLines(string(data))
	if len(lines) == 0 {
		return nil
	}
	lines = lines[:len(lines)-1]
	return os.WriteFile(path, []byte(joinLines(lines)), 0o644)
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
