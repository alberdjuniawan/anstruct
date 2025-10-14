package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alberdjuniawan/anstruct/internal/core"
	"github.com/spf13/cobra"
)

func newMStructCmd() *cobra.Command {
	var (
		outDir        string
		dry           bool
		force         bool
		verbose       bool
		allowReserved bool
	)

	cmd := &cobra.Command{
		Use:   "mstruct <file.struct>",
		Short: "Generate project files from a .struct blueprint",
		Long: `mstruct reads a .struct blueprint and generates directories/files
based on its structure definition.

Examples:
  anstruct mstruct myapp.struct
  anstruct mstruct -o ./generated myapp.struct
  anstruct mstruct --force ./blueprints/web.struct
  anstruct mstruct --dry --verbose ./blueprints/api.struct
  anstruct mstruct --allow-reserved myapp.struct  # include vendor/, node_modules/`,
		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			structFile := filepath.Clean(args[0])

			info, err := os.Stat(structFile)
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", structFile)
			}
			if info.IsDir() {
				return fmt.Errorf("expected a .struct file, got a directory: %s", structFile)
			}
			if filepath.Ext(structFile) != ".struct" {
				return fmt.Errorf("invalid file type: %s (must be .struct)", structFile)
			}

			cleanOutDir := filepath.Clean(outDir)
			if _, err := os.Stat(cleanOutDir); os.IsNotExist(err) {
				if mkErr := os.MkdirAll(cleanOutDir, 0755); mkErr != nil {
					return fmt.Errorf("failed to create output directory: %w", mkErr)
				}
			}

			fmt.Printf("🚧 Generating project from %s → %s\n", structFile, cleanOutDir)
			if dry {
				fmt.Println("💡 Dry run mode enabled: no files will be written.")
			}
			if allowReserved {
				fmt.Println("⚠️  --allow-reserved enabled: reserved folders will be included")
			}

			receipt, err := svc.MStruct(ctx, structFile, cleanOutDir, core.GenerateOptions{
				DryRun:        dry,
				Force:         force,
				AllowReserved: allowReserved,
			})
			if err != nil {
				return fmt.Errorf("generation failed: %w", err)
			}

			if dry && verbose {
				fmt.Println("\n📂 Preview of what would be generated:")
				if len(receipt.CreatedDirs) > 0 {
					fmt.Println("  📁 Directories:")
					for _, dir := range receipt.CreatedDirs {
						fmt.Printf("    - %s\n", dir)
					}
				}
				if len(receipt.CreatedFiles) > 0 {
					fmt.Println("  📄 Files:")
					for _, file := range receipt.CreatedFiles {
						fmt.Printf("    - %s\n", file)
					}
				}
			}

			fmt.Printf("\n✅ Done! %d directories, %d files created.\n",
				len(receipt.CreatedDirs), len(receipt.CreatedFiles))

			if dry {
				fmt.Println("📍 (Dry run completed — no actual files written.)")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outDir, "out", "o", ".", "output directory (default: current folder)")
	cmd.Flags().BoolVar(&dry, "dry", false, "simulate generation without writing files")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files if they already exist")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed preview of generated structure")
	cmd.Flags().BoolVar(&allowReserved, "allow-reserved", false, "allow reserved folders like vendor/, node_modules/ (not recommended)")

	return cmd
}
