package watcher

import (
	"context"
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
}

type Watcher struct{}

func New() *Watcher { return &Watcher{} }

func (w *Watcher) Run(ctx context.Context, cfg SyncConfig, onFolder func(), onBlueprint func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// --- recursive watch untuk semua subdir ---
	err = filepath.WalkDir(cfg.ProjectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
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

	// watch file blueprint
	if err := watcher.Add(cfg.BlueprintPath); err != nil {
		return err
	}

	// --- debounce safe ---
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

	// --- event loop ---
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watcher.Events:
			// filter event
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				if sameFile(event.Name, cfg.BlueprintPath) {
					trigger("blueprint")
				} else if strings.HasPrefix(event.Name, cfg.ProjectPath) {
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
