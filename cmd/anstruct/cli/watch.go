package cli

import (
	"context"
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
	var (
		halfMode string
		fullMode bool
		dry      bool
		verbose  bool
		ignore   string
		debounce time.Duration
	)

	cmd := &cobra.Command{
		Use:   "watch [projectDir] [blueprintFile]",
		Short: "Watch and sync project ↔ blueprint in real time",
		Long: `Synchronize changes between your project folder and .struct blueprint.

Modes:
  --half struct   : sync only from blueprint → folder
  --half folder   : sync only from folder → blueprint
  --full          : enable full two-way sync
Flags:
  --dry           : simulate without writing files
  --verbose, -v   : print detailed file changes
  --ignore <patt> : skip files or dirs matching pattern
  --debounce <d>  : set delay before reacting (default 2s)

Examples:
  anstruct watch ./myapp ./myapp.struct --half folder
  anstruct watch ./myapp ./myapp.struct --half struct --verbose
  anstruct watch ./myapp ./myapp.struct --full --ignore node_modules --debounce 1s
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			project := filepath.Clean(args[0])
			blueprint := filepath.Clean(args[1])

			cfg := watcher.SyncConfig{
				ProjectPath:   project,
				BlueprintPath: blueprint,
				Debounce:      debounce,
				Verbose:       verbose,
				IgnorePattern: ignore,
			}

			cmd.Printf("👀 Watching project: %s\n📜 Blueprint: %s\n", project, blueprint)
			cmd.Printf("⚙️  Mode: %s | Dry: %v | Ignore: %s | Debounce: %v\n",
				modeLabel(halfMode, fullMode), dry, ignore, debounce)

			// --- define actions ---
			onFolder := func() {
				if dry {
					cmd.Println("📂 (dry-run) Folder changed → would update blueprint")
					return
				}
				cmd.Println("📂 Folder changed → updating blueprint")
				if err := svc.RStruct(ctx, project, blueprint); err != nil {
					cmd.Printf("RStruct error: %v\n", err)
				}
			}

			onBlueprint := func() {
				if dry {
					cmd.Println("📜 (dry-run) Blueprint changed → would regenerate project")
					return
				}
				cmd.Println("📜 Blueprint changed → regenerating project")
				tree, err := svc.Parser.Parse(ctx, blueprint)
				if err != nil {
					cmd.Printf("Parse error: %v\n", err)
					return
				}

				receipt, err := svc.Writer.Generate(ctx, tree, project, core.GenerateOptions{Force: true})
				if err != nil {
					cmd.Printf("Generate error: %v\n", err)
					return
				}

				allowed := map[string]bool{}
				for _, c := range tree.Root.Children {
					generator.CollectAllowed(c, "", allowed)
				}
				addReservedAllowed(project, blueprint, allowed)
				if err := generator.CleanupExtra(project, allowed); err != nil {
					cmd.Printf("Cleanup error: %v\n", err)
				}

				cmd.Printf("✅ Synced: %d dirs, %d files\n", receipt.CreatedDirs, receipt.CreatedFiles)
			}

			// --- mode handling ---
			switch {
			case fullMode:
				cmd.Println("🔁 Running in FULL sync mode (bi-directional)")
				return watcher.New().Run(cmd.Context(), cfg, onFolder, onBlueprint)

			case halfMode == "folder":
				cmd.Println("↩️ Running in HALF mode (folder → struct)")
				return watcher.New().Run(cmd.Context(), cfg, onFolder, nil)

			case halfMode == "struct":
				cmd.Println("➡️ Running in HALF mode (struct → folder)")
				return watcher.New().Run(cmd.Context(), cfg, nil, onBlueprint)

			default:
				cmd.Println("⚠️ Must specify --half (folder|struct) or --full")
				return cmd.Help()
			}
		},
	}

	cmd.Flags().StringVar(&halfMode, "half", "", "run in one-way sync mode: 'folder' or 'struct'")
	cmd.Flags().BoolVar(&fullMode, "full", false, "enable two-way full sync mode")
	cmd.Flags().BoolVar(&dry, "dry", false, "simulate without modifying any files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	cmd.Flags().StringVar(&ignore, "ignore", "", "ignore files or dirs matching pattern")
	cmd.Flags().DurationVar(&debounce, "debounce", 2*time.Second, "debounce interval for file change detection")

	return cmd
}

func modeLabel(half string, full bool) string {
	if full {
		return "full"
	}
	if half != "" {
		return "half-" + half
	}
	return "unknown"
}

// --- Helper: reserved protection ---
func addReservedAllowed(project string, blueprint string, allowed map[string]bool) {
	projectAbs, _ := filepath.Abs(project)
	blueprintAbs, _ := filepath.Abs(blueprint)

	if isSubPath(blueprintAbs, projectAbs) {
		rel, err := filepath.Rel(projectAbs, blueprintAbs)
		if err == nil {
			markAllowedPath(rel, allowed)
		}
	}

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
