package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newAIStructCmd() *cobra.Command {
	var outFile string

	cmd := &cobra.Command{
		Use:   "aistruct <prompt>",
		Short: "Generate .struct blueprint using AI prompt",
		Args:  cobra.MinimumNArgs(1), // ✅ boleh banyak argumen
		RunE: func(cmd *cobra.Command, args []string) error {
			// Gabungkan semua argumen jadi satu kalimat
			prompt := strings.Join(args, " ")

			// Default output jika user tidak pakai --out
			if outFile == "" {
				outFile = "aistruct.struct"
			}

			if err := svc.AIStruct(ctx, prompt, outFile); err != nil {
				return fmt.Errorf("AIStruct error: %w", err)
			}

			fmt.Println("✅ Blueprint written to", outFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outFile, "out", "o", "", "output .struct file path")
	return cmd
}
