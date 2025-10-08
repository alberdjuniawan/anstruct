package history

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
		return err
	}
	lines := splitLines(string(data))
	if len(lines) == 0 {
		return errors.New("no history to undo")
	}
	last := lines[len(lines)-1]
	var op core.Operation
	if err := json.Unmarshal([]byte(last), &op); err != nil {
		return err
	}

	switch op.Type {
	case core.OpCreate:
		// rollback: hapus file dulu, lalu folder (bottom-up)
		for _, f := range op.Receipt.CreatedFiles {
			_ = os.Remove(f)
		}
		// urutkan folder dari dalam ke luar
		dirs := op.Receipt.CreatedDirs
		sort.Slice(dirs, func(i, j int) bool {
			return len(dirs[i]) > len(dirs[j])
		})
		for _, d := range dirs {
			_ = os.Remove(d)
		}
	case core.OpReverse:
		_ = os.Remove(op.Target) // hapus blueprint hasil reverse
	default:
		return errors.New("unsupported undo type: " + string(op.Type))
	}

	// hapus baris terakhir dari log
	return truncateLastLine(h.LogPath)
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
	return filepath.SplitList(strings.ReplaceAll(s, "\r\n", "\n"))
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
