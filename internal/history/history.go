package history

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type History struct {
	LogPath       string
	UndoStackPath string
	Recreator     OperationRecreator
}

type OperationRecreator interface {
	RecreateOperation(ctx context.Context, op core.Operation) error
}

func New(logPath string) *History {
	dir := filepath.Dir(logPath)
	return &History{
		LogPath:       logPath,
		UndoStackPath: filepath.Join(dir, "undo_stack.log"),
	}
}

func (h *History) SetRecreator(r OperationRecreator) {
	h.Recreator = r
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

	if err := h.clearUndoStack(); err != nil {
		fmt.Printf("⚠️  Warning: Failed to clear undo stack: %v\n", err)
	}

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
		return fmt.Errorf("failed to parse operation: %w", err)
	}

	if err := h.undoOperation(op); err != nil {
		return fmt.Errorf("undo operation failed: %w", err)
	}

	if err := h.pushToUndoStack(op); err != nil {
		return fmt.Errorf("failed to save to undo stack: %w", err)
	}

	if err := truncateLastLine(h.LogPath); err != nil {
		return fmt.Errorf("failed to update history: %w", err)
	}

	return nil
}

func (h *History) Redo(ctx context.Context) error {
	data, err := os.ReadFile(h.UndoStackPath)
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
		return fmt.Errorf("failed to parse operation: %w", err)
	}

	if h.Recreator != nil {
		if err := h.Recreator.RecreateOperation(ctx, op); err != nil {
			return fmt.Errorf("failed to recreate operation: %w", err)
		}
	} else {
		return fmt.Errorf("cannot redo: recreator not set")
	}

	if err := truncateLastLine(h.UndoStackPath); err != nil {
		return fmt.Errorf("failed to update undo stack: %w", err)
	}

	if err := h.recordWithoutClearingUndoStack(ctx, op); err != nil {
		return fmt.Errorf("failed to restore to history: %w", err)
	}

	return nil
}

func (h *History) recordWithoutClearingUndoStack(ctx context.Context, op core.Operation) error {
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

func (h *History) undoOperation(op core.Operation) error {
	switch op.Type {
	case core.OpCreate, core.OpAIApply:
		return h.undoCreate(op)

	case core.OpReverse:
		return h.undoReverse(op)

	case core.OpAI:
		return h.undoAIBlueprint(op)

	default:
		return fmt.Errorf("unsupported undo type: %s", op.Type)
	}
}

func (h *History) undoCreate(op core.Operation) error {
	var errors []string

	for _, f := range op.Receipt.CreatedFiles {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("file %s: %v", f, err))
		}
	}

	dirs := op.Receipt.CreatedDirs
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	for _, d := range dirs {
		if err := os.Remove(d); err != nil && !os.IsNotExist(err) {
			continue
		}
	}

	if len(errors) > 0 {
		fmt.Printf("⚠️  Some files could not be removed:\n")
		for _, e := range errors {
			fmt.Printf("   - %s\n", e)
		}
	}

	return nil
}

func (h *History) undoReverse(op core.Operation) error {
	if err := os.Remove(op.Target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove reversed blueprint %s: %w", op.Target, err)
	}
	return nil
}

func (h *History) undoAIBlueprint(op core.Operation) error {
	if err := os.Remove(op.Target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove AI blueprint %s: %w", op.Target, err)
	}
	return nil
}

func (h *History) pushToUndoStack(op core.Operation) error {
	_ = os.MkdirAll(filepath.Dir(h.UndoStackPath), 0o755)
	f, err := os.OpenFile(h.UndoStackPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(op)
}

func (h *History) clearUndoStack() error {
	if _, err := os.Stat(h.UndoStackPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(h.UndoStackPath)
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

func (h *History) ListUndoStack(ctx context.Context) ([]core.Operation, error) {
	data, err := os.ReadFile(h.UndoStackPath)
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
	_ = os.Remove(h.LogPath)
	_ = os.Remove(h.UndoStackPath)
	return nil
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
