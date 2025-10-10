package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type SyncConfig struct {
	ProjectPath   string        // folder proyek
	BlueprintPath string        // file .struct
	Debounce      time.Duration // jeda sebelum trigger
	Verbose       bool
	IgnorePattern string // pattern yg di-skip
}

type Watcher struct{}

func New() *Watcher { return &Watcher{} }

func (w *Watcher) Run(ctx context.Context, cfg SyncConfig, onFolder func(), onBlueprint func()) error {
	projectAbs, err := filepath.Abs(cfg.ProjectPath)
	if err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}
	blueprintAbs, err := filepath.Abs(cfg.BlueprintPath)
	if err != nil {
		return fmt.Errorf("invalid blueprint path: %w", err)
	}
	if _, err := os.Stat(blueprintAbs); err != nil {
		return fmt.Errorf("blueprint file not found: %s", blueprintAbs)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// watch recursive
	err = filepath.WalkDir(projectAbs, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldIgnore(cfg.IgnorePattern, path) {
				return filepath.SkipDir
			}
			if cfg.Verbose {
				log.Println("watching dir:", path)
			}
			if err := watcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if cfg.Verbose {
		log.Println("watching blueprint:", blueprintAbs)
	}
	if err := watcher.Add(blueprintAbs); err != nil {
		return err
	}

	var mu sync.Mutex
	var timer *time.Timer
	trigger := func(source string) {
		mu.Lock()
		defer mu.Unlock()
		if cfg.Verbose {
			log.Println("change detected from:", source)
		}
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(cfg.Debounce, func() {
			if cfg.Verbose {
				log.Println("sync triggered by:", source)
			}
			if source == "blueprint" && onBlueprint != nil {
				onBlueprint()
			}
			if source == "folder" && onFolder != nil {
				onFolder()
			}
		})
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				if shouldIgnore(cfg.IgnorePattern, event.Name) {
					continue
				}
				if sameFile(event.Name, blueprintAbs) {
					trigger("blueprint")
				} else if strings.HasPrefix(event.Name, projectAbs) {
					trigger("folder")
				}
			}
		case err := <-watcher.Errors:
			log.Println("watcher error:", err)
		}
	}
}

// --- helpers ---
func sameFile(a, b string) bool {
	ap, _ := filepath.Abs(a)
	bp, _ := filepath.Abs(b)
	return ap == bp
}

func shouldIgnore(pattern, path string) bool {
	if pattern == "" {
		return false
	}
	return strings.Contains(path, pattern)
}
