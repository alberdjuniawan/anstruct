package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newRStructCmd() *cobra.Command {
	var (
		outFile string
		dry     bool
		verbose bool
	)

	cmd := &cobra.Command{
		Use:   "rstruct <projectDir>",
		Short: "Reverse a project folder into a .struct blueprint",
		Long: `rstruct scans a project directory and generates a .struct blueprint
representation of its structure.

Examples:
  anstruct rstruct ./myapp
  anstruct rstruct -o ./blueprints/app.struct ./projects/web
  anstruct rstruct -o ./blueprints ./myapp
  anstruct rstruct --dry ./examples/demo
  anstruct rstruct --verbose ./api`,

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := filepath.Clean(args[0])

			info, err := os.Stat(projectDir)
			if os.IsNotExist(err) {
				return fmt.Errorf("directory not found: %s", projectDir)
			}
			if !info.IsDir() {
				return fmt.Errorf("expected a directory, got a file: %s", projectDir)
			}

			outFile = resolveOutputPath(outFile, projectDir)

			outDir := filepath.Dir(outFile)
			if _, err := os.Stat(outDir); os.IsNotExist(err) {
				if mkErr := os.MkdirAll(outDir, 0755); mkErr != nil {
					return fmt.Errorf("failed to create output dir: %w", mkErr)
				}
			}

			fmt.Printf("🔄 Reversing project from %s → %s\n", projectDir, outFile)
			if dry {
				fmt.Println("💡 Dry run mode enabled: no files will be written.")
			}

			if dry {
				fmt.Println("🔍 (Dry run) Listing structure...")
				printDirTree(projectDir, verbose)
				fmt.Printf("\n✅ Dry run complete. Blueprint would be written to: %s\n", outFile)
				return nil
			}

			if err := svc.RStruct(ctx, projectDir, outFile); err != nil {
				return fmt.Errorf("RStruct error: %w", err)
			}

			fmt.Printf("\n✅ Done! Blueprint written to %s\n", outFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outFile, "out", "o", "", "output .struct file or directory (auto adds .struct if missing)")
	cmd.Flags().BoolVar(&dry, "dry", false, "simulate reverse without writing .struct file")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed directory tree (used with --dry)")

	return cmd
}

func printDirTree(root string, verbose bool) {
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil
		}
		indent := strings.Repeat("  ", strings.Count(rel, string(os.PathSeparator)))
		if d.IsDir() {
			fmt.Printf("%s📁 %s\n", indent, d.Name())
		} else if verbose {
			fmt.Printf("%s📄 %s\n", indent, d.Name())
		}
		return nil
	})
}

func resolveOutputPath(outArg, projectDir string) string {
	base := filepath.Base(projectDir)

	if outArg == "" {
		return fmt.Sprintf("%s.struct", base)
	}

	clean := filepath.Clean(outArg)
	if strings.HasSuffix(clean, ".struct") {
		return clean
	}

	if strings.HasSuffix(outArg, "/") || strings.HasSuffix(outArg, "\\") {
		return filepath.Join(clean, fmt.Sprintf("%s.struct", base))
	}

	if info, err := os.Stat(clean); err == nil && info.IsDir() {
		return filepath.Join(clean, fmt.Sprintf("%s.struct", base))
	}

	if filepath.Ext(clean) == "" {
		return fmt.Sprintf("%s.struct", clean)
	}

	return clean
}
