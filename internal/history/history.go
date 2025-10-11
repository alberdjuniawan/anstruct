package history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (h *History) Record(ctx context.Context, op core.Operation) error {
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

	if err := h.undoOperation(op); err != nil {
		return err
	}

	return truncateLastLine(h.LogPath)
}

func (h *History) undoOperation(op core.Operation) error {
	switch op.Type {
	case core.OpCreate:
		return h.undoCreate(op)

	case core.OpReverse:
		return h.undoReverse(op)

	case core.OpAI:
		return h.undoAIBlueprint(op)

	case core.OpAIApply:
		return h.undoAIApply(op)

	default:
		return errors.New("unsupported undo type: " + string(op.Type))
	}
}

func (h *History) undoCreate(op core.Operation) error {
	for _, f := range op.Receipt.CreatedFiles {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Failed to remove file %s: %v\n", f, err)
		}
	}

	dirs := op.Receipt.CreatedDirs
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	for _, d := range dirs {
		if err := os.Remove(d); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Failed to remove directory %s: %v\n", d, err)
		}
	}

	return nil
}

func (h *History) undoReverse(op core.Operation) error {
	return os.Remove(op.Target)
}

func (h *History) undoAIBlueprint(op core.Operation) error {
	return os.Remove(op.Target)
}

func (h *History) undoAIApply(op core.Operation) error {
	return h.undoCreate(op)
}

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
			continue
		}
		ops = append(ops, op)
	}

	return ops, nil
}

func (h *History) Clear(ctx context.Context) error {
	return os.Remove(h.LogPath)
}

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
