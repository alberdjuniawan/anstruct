package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/alberdjuniawan/anstruct/internal/ai"
	"github.com/alberdjuniawan/anstruct/internal/converter"
	"github.com/spf13/cobra"
)

func newConvertCmd() *cobra.Command {
	var (
		outFile string
		format  string
		mode    string
		stdin   bool
		verbose bool
	)

	cmd := &cobra.Command{
		Use:   "convert <input-file>",
		Short: "Convert various structure formats to .struct",
		Long: `Convert different project structure formats to anstruct .struct format.

Supported input formats:
  - tree      : tree command output (├──, └──, │)
  - ls        : ls -R output
  - markdown  : markdown with tree symbols
  - plain     : plain indented text
  - auto      : auto-detect format (default)

Normalization modes:
  - auto      : Try AI first, fallback to manual (default)
  - ai        : AI-powered normalization only (best for messy formats)
  - manual    : Regex-based parsing only (offline mode)
  - offline   : Alias for manual

Examples:
  # From tree command output
  tree myproject > structure.txt
  anstruct convert structure.txt -o myproject.struct

  # From stdin
  tree myproject | anstruct convert --stdin -o myproject.struct

  # With AI normalization (auto-clean messy formats)
  anstruct convert messy-structure.txt --mode ai -o clean.struct

  # Offline mode (no AI, faster)
  anstruct convert simple-tree.txt --mode offline -o output.struct

  # Auto mode (recommended)
  anstruct convert random-format.txt --mode auto --verbose`,

		Args: func(cmd *cobra.Command, args []string) error {
			if stdin {
				return nil
			}
			if len(args) < 1 {
				return fmt.Errorf("requires input file or --stdin flag")
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			var input string
			var inputSource string

			// Read input
			if stdin {
				inputSource = "stdin"
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				input = string(data)
			} else {
				inputSource = args[0]
				data, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				input = string(data)
			}

			// Default output
			if outFile == "" {
				outFile = "converted.struct"
			}

			fmt.Printf("🔄 Converting structure from %s...\n", inputSource)
			if verbose {
				fmt.Printf("📊 Input size: %d bytes\n", len(input))
				fmt.Printf("🎯 Mode: %s\n", mode)
			}

			// Create converter based on mode
			var conv *converter.Converter
			normMode := converter.NormalizationMode(mode)

			switch normMode {
			case converter.ModeAI, converter.ModeAuto:
				// Create with AI support
				provider := ai.NewGeminiProvider("")
				conv = converter.NewWithAI(provider, svc.Parser, normMode)
				if verbose {
					fmt.Println("🤖 AI mode enabled")
				}
			case converter.ModeManual, converter.ModeOffline:
				// Manual mode (no AI)
				conv = converter.New()
				if verbose {
					fmt.Println("⚙️  Manual mode (offline)")
				}
			default:
				return fmt.Errorf("invalid mode: %s (use: auto, ai, manual, offline)", mode)
			}

			// Convert
			tree, detectedFormat, err := conv.Convert(ctx, input)
			if err != nil {
				return fmt.Errorf("conversion failed: %w", err)
			}

			fmt.Printf("✅ Detected format: %s\n", detectedFormat)

			if verbose {
				fmt.Printf("📂 Root: %s\n", tree.Root.Name)
				fmt.Printf("📊 Children: %d items\n", len(tree.Root.Children))
			}

			// Write to .struct file
			output := conv.ConvertToString(tree)

			if err := os.WriteFile(outFile, []byte(output), 0644); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}

			fmt.Printf("💾 Saved to: %s\n", outFile)

			if verbose {
				fmt.Println("\n📄 Preview:")
				fmt.Println("---")
				lines := splitPreview(output, 20)
				for _, line := range lines {
					fmt.Println(line)
				}
				if len(output) > 1000 {
					fmt.Println("... (truncated)")
				}
				fmt.Println("---")
			}

			fmt.Println("\n✅ Conversion completed successfully!")
			fmt.Println("💡 Next steps:")
			fmt.Printf("   anstruct mstruct %s          # Generate project\n", outFile)
			fmt.Printf("   cat %s                       # View blueprint\n", outFile)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outFile, "out", "o", "", "output .struct file (default: converted.struct)")
	cmd.Flags().StringVar(&format, "format", "auto", "input format (auto, tree, ls, markdown, plain)")
	cmd.Flags().StringVar(&mode, "mode", "auto", "normalization mode (auto, ai, manual, offline)")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "read from stdin instead of file")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed conversion info")

	return cmd
}

// splitPreview splits output into lines for preview
func splitPreview(s string, maxLines int) []string {
	lines := []string{}
	current := ""
	count := 0

	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
			count++
			if count >= maxLines {
				break
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" && count < maxLines {
		lines = append(lines, current)
	}

	return lines
}
