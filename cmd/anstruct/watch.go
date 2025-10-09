package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alberdjuniawan/anstruct"
	"github.com/alberdjuniawan/anstruct/internal/core"
	"github.com/alberdjuniawan/anstruct/internal/generator"
	"github.com/alberdjuniawan/anstruct/internal/watcher"
)

func newWatchCmd(svc *anstruct.Service) *cobra.Command {
	var verbose bool
	var debounce time.Duration

	cmd := &cobra.Command{
		Use:   "watch [projectDir] [blueprintFile]",
		Short: "Watch project folder and blueprint file for changes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := filepath.Clean(args[0])
			blueprint := filepath.Clean(args[1])

			cfg := watcher.SyncConfig{
				ProjectPath:   project,
				BlueprintPath: blueprint,
				Debounce:      debounce,
				Verbose:       verbose,
			}

			fmt.Printf("👀 Watching project: %s and blueprint: %s\n", project, blueprint)

			onFolder := func() {
				fmt.Println("📂 Folder changed → update blueprint")
				if err := svc.RStruct(context.Background(), project, blueprint); err != nil {
					fmt.Println("RStruct error:", err)
				}
			}

			onBlueprint := func() {
				fmt.Println("📜 Blueprint changed → sync folder")
				tree, err := svc.Parser.Parse(context.Background(), blueprint)
				if err != nil {
					fmt.Println("Parse error:", err)
					return
				}

				receipt, err := svc.Writer.Generate(context.Background(), tree, project, core.GenerateOptions{Force: true})
				if err != nil {
					fmt.Println("Generate error:", err)
					return
				}

				allowed := map[string]bool{}
				for _, c := range tree.Root.Children {
					generator.CollectAllowed(c, "", allowed)
				}

				addReservedAllowed(project, blueprint, allowed)

				if err := generator.CleanupExtra(project, allowed); err != nil {
					fmt.Println("Cleanup error:", err)
				}

				fmt.Println("✅ Synced:", receipt.CreatedDirs, receipt.CreatedFiles)
			}

			return watcher.New().Run(cmd.Context(), cfg, onFolder, onBlueprint)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	cmd.Flags().DurationVar(&debounce, "debounce", 2*time.Second, "debounce interval")

	return cmd
}

// --- Reserved protection ---

func addReservedAllowed(project string, blueprint string, allowed map[string]bool) {
	projectAbs, _ := filepath.Abs(project)
	blueprintAbs, _ := filepath.Abs(blueprint)

	// Protect blueprint file if inside project
	if isSubPath(blueprintAbs, projectAbs) {
		rel, err := filepath.Rel(projectAbs, blueprintAbs)
		if err == nil {
			markAllowedPath(rel, allowed)
		}
	}

	// Walk project and protect .struct files + reserved folders
	_ = filepath.WalkDir(projectAbs, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && isReservedDir(d.Name()) {
			rel, rerr := filepath.Rel(projectAbs, path)
			if rerr == nil {
				markAllowedPath(rel, allowed)
			}
			return filepath.SkipDir
		}
		if !d.IsDir() && hasStructSuffix(d.Name()) {
			rel, rerr := filepath.Rel(projectAbs, path)
			if rerr == nil {
				markAllowedPath(rel, allowed)
			}
		}
		return nil
	})
}

func isReservedDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor":
		return true
	default:
		return false
	}
}

func hasStructSuffix(name string) bool {
	return filepath.Ext(name) == ".struct"
}

func isSubPath(childAbs string, parentAbs string) bool {
	rel, err := filepath.Rel(parentAbs, childAbs)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel)
}

func markAllowedPath(rel string, allowed map[string]bool) {
	key := filepath.ToSlash(rel)
	allowed[key] = true
}
