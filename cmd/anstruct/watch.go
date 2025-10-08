package main

import (
	"context"
	"fmt"
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
			project := args[0]
			blueprint := args[1]

			cfg := watcher.SyncConfig{
				ProjectPath:   project,
				BlueprintPath: blueprint,
				Debounce:      debounce,
				Verbose:       verbose,
			}

			fmt.Printf("👀 Watching project: %s and blueprint: %s\n", project, blueprint)

			onFolder := func() {
				fmt.Println("📂 Folder changed → update blueprint")
				if _, err := svc.RStruct(context.Background(), project); err != nil {
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
				// full sync: overwrite + cleanup
				receipt, err := svc.Writer.Generate(context.Background(), tree, project, core.GenerateOptions{Force: true})
				if err != nil {
					fmt.Println("Generate error:", err)
					return
				}
				allowed := map[string]bool{}
				for _, c := range tree.Root.Children {
					generator.CollectAllowed(c, "", allowed)
				}
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
