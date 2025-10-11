package cli

import (
	"context"
	"fmt"

	"github.com/alberdjuniawan/anstruct/internal/core"
	"github.com/spf13/cobra"
)

func newUndoCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo the last operation",
		Long: `Undo the last anstruct operation (aistruct, mstruct, or rstruct).

This will:
- Delete files/folders created by the last operation
- Remove the operation from history
- Cannot be undone itself

Examples:
  anstruct undo
  anstruct undo --confirm`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				fmt.Println("⚠️  WARNING: This will delete files created by the last operation!")
				fmt.Println("Use --confirm flag to proceed, or use 'anstruct history' to see what will be undone.")
				return nil
			}

			if err := svc.History.Undo(ctx); err != nil {
				return fmt.Errorf("undo failed: %w", err)
			}

			fmt.Println("✅ Last operation undone successfully!")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&confirm, "confirm", "y", false, "confirm undo without prompt")

	return cmd
}

func newHistoryCmd() *cobra.Command {
	var clear bool

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show operation history",
		Long: `Display the history of all anstruct operations.

Each entry shows:
- Operation type (aistruct, mstruct, rstruct)
- Target path
- Timestamp
- Created files/directories count

Examples:
  anstruct history
  anstruct history --clear`,

		RunE: func(cmd *cobra.Command, args []string) error {
			histImpl, ok := svc.History.(interface {
				List(ctx context.Context) ([]core.Operation, error)
				Clear(ctx context.Context) error
			})

			if !ok {
				return fmt.Errorf("history implementation does not support list/clear")
			}

			if clear {
				if err := histImpl.Clear(ctx); err != nil {
					return fmt.Errorf("failed to clear history: %w", err)
				}
				fmt.Println("✅ History cleared")
				return nil
			}

			ops, err := histImpl.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to read history: %w", err)
			}

			if len(ops) == 0 {
				fmt.Println("📜 No operations in history")
				return nil
			}

			fmt.Printf("📜 Operation History (%d entries):\n\n", len(ops))

			for i, op := range ops {
				fmt.Printf("%d. [%s] %s\n", i+1, op.Type, op.Timestamp)
				fmt.Printf("   Target: %s\n", op.Target)

				if len(op.Receipt.CreatedDirs) > 0 || len(op.Receipt.CreatedFiles) > 0 {
					fmt.Printf("   Created: %d dirs, %d files\n",
						len(op.Receipt.CreatedDirs),
						len(op.Receipt.CreatedFiles))
				}
				fmt.Println()
			}

			fmt.Println("💡 Use 'anstruct undo --confirm' to undo the last operation")
			return nil
		},
	}

	cmd.Flags().BoolVar(&clear, "clear", false, "clear all history")

	return cmd
}
