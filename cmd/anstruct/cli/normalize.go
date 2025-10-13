package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alberdjuniawan/anstruct/internal/core"
	"github.com/spf13/cobra"
)

func newNormalizeCmd() *cobra.Command {
	var (
		outFile string
		dry     bool
		verbose bool
		retries int
	)

	cmd := &cobra.Command{
		Use:   "normalize <input-file>",
		Short: "Convert any text-based structure format to .struct format",
		Long: `Normalize converts various project structure formats into the standard .struct format.

Supported input formats:
  • Tree command output (tree -L 3)
  • ASCII art structures
  • Markdown directory listings
  • Plain text with indentation
  • Any text-based structure representation

The AI will intelligently parse and convert it to proper .struct format.

Examples:
  # From tree command output
  tree -L 3 > structure.txt
  anstruct normalize structure.txt -o myproject.struct

  # From markdown file
  anstruct normalize README.md -o structure.struct

  # From any text file
  anstruct normalize notes.txt -o normalized.struct --verbose

  # Preview without saving
  anstruct normalize input.txt --dry`,

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := filepath.Clean(args[0])

			if _, err := os.Stat(inputFile); os.IsNotExist(err) {
				return fmt.Errorf("input file not found: %s", inputFile)
			}

			content, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}

			inputContent := string(content)
			if strings.TrimSpace(inputContent) == "" {
				return fmt.Errorf("input file is empty")
			}

			if outFile == "" {
				baseName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
				outFile = baseName + ".struct"
			}

			if !strings.HasSuffix(outFile, ".struct") {
				outFile = outFile + ".struct"
			}

			fmt.Printf("📥 Input: %s\n", inputFile)
			fmt.Printf("📤 Output: %s\n", outFile)

			if dry {
				fmt.Println("💡 Dry run mode enabled: no files will be written.")
			}

			opts := core.AIOptions{
				DryRun:  dry,
				Verbose: verbose,
				Retries: retries,
			}

			if err := svc.NormalizeStruct(ctx, inputContent, outFile, opts); err != nil {
				return fmt.Errorf("normalization failed: %w", err)
			}

			if !dry {
				fmt.Println("\n✅ Structure normalized successfully!")
				fmt.Printf("💡 You can now use: anstruct mstruct %s\n", outFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outFile, "out", "o", "", "output .struct file (default: <input-name>.struct)")
	cmd.Flags().BoolVar(&dry, "dry", false, "preview normalized structure without writing")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show AI processing details")
	cmd.Flags().IntVar(&retries, "retries", 2, "retry count if AI output invalid")

	return cmd
}
