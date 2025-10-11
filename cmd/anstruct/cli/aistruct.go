package cli

import (
	"fmt"
	"strings"

	"github.com/alberdjuniawan/anstruct/internal/core"
	"github.com/spf13/cobra"
)

func newAIStructCmd() *cobra.Command {
	var (
		outFile string
		apply   bool
		dry     bool
		verbose bool
		retries int
		force   bool
	)

	cmd := &cobra.Command{
		Use:   "aistruct <prompt>",
		Short: "Generate .struct blueprint or folder using AI",
		Long: `Generate project structure from natural language using AI.

Modes:
  Default: Generate .struct blueprint file
  --apply: Generate project folder directly from AI result

Examples:
  anstruct aistruct "flutter app with auth" -o ./myapp.struct
  anstruct aistruct "nodejs api with routes" --apply -o ./myapi
  anstruct aistruct "react dashboard" --dry --verbose
  anstruct aistruct "golang microservice" --apply --force`,

		Args: cobra.MinimumNArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := strings.Join(args, " ")

			if outFile == "" {
				if apply {
					outFile = "./aiproject"
				} else {
					outFile = "aistruct.struct"
				}
			}

			isStructOutput := strings.HasSuffix(outFile, ".struct")
			if !apply && !isStructOutput && !dry {
				return fmt.Errorf("output must end with .struct or use --apply to generate folder")
			}

			opts := core.AIOptions{
				Apply:   apply,
				DryRun:  dry,
				Verbose: verbose,
				Retries: retries,
				Force:   force,
			}

			if apply {
				fmt.Printf("🤖 AI Mode: Generate folder directly → %s\n", outFile)
			} else {
				fmt.Printf("🤖 AI Mode: Generate blueprint → %s\n", outFile)
			}

			if dry {
				fmt.Println("💡 Dry run mode enabled: no files will be written.")
			}

			if err := svc.AIStruct(ctx, prompt, outFile, opts); err != nil {
				return fmt.Errorf("AIStruct error: %w", err)
			}

			fmt.Println("✅ AIstruct completed successfully.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&outFile, "out", "o", "", "output path (.struct file or folder)")
	cmd.Flags().BoolVar(&apply, "apply", false, "generate folder directly from AI result")
	cmd.Flags().BoolVar(&dry, "dry", false, "simulate without writing")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show raw AI output")
	cmd.Flags().IntVar(&retries, "retries", 2, "retry count if AI output invalid")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files when using --apply")

	return cmd
}
