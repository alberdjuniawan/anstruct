package cli

import (
	"fmt"

	"github.com/alberdjuniawan/anstruct/internal/core"
	"github.com/alberdjuniawan/anstruct/internal/history"
	"github.com/spf13/cobra"
)

// newHistoryCmd returns the main history command with subcommands
func newHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Manage operation history (list, undo, redo, clear)",
		Long: `Manage the history of all anstruct operations.

Available subcommands:
  list   - Show all operations
  undo   - Undo last operation
  redo   - Redo last undone operation
  clear  - Clear all history

Examples:
  anstruct history list
  anstruct history undo
  anstruct history redo
  anstruct history clear`,
	}

	// Add subcommands
	cmd.AddCommand(
		newHistoryListCmd(),
		newHistoryUndoCmd(),
		newHistoryRedoCmd(),
		newHistoryClearCmd(),
	)

	return cmd
}

// newHistoryListCmd shows operation history
func newHistoryListCmd() *cobra.Command {
	var showUndoStack bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show all operations in history",
		Long: `Display the history of all anstruct operations with details.

Each entry shows:
- Operation type (aistruct, mstruct, rstruct)
- Target path
- Timestamp
- Created files/directories count

Examples:
  anstruct history list
  anstruct history list --undo-stack  # Show operations that can be redone`,

		RunE: func(cmd *cobra.Command, args []string) error {
			histImpl, ok := svc.History.(*history.History)
			if !ok {
				return fmt.Errorf("history implementation error")
			}

			if showUndoStack {
				return displayUndoStack(histImpl)
			}

			ops, err := histImpl.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to read history: %w", err)
			}

			if len(ops) == 0 {
				fmt.Println("📜 No operations in history")
				fmt.Println("\n💡 Operations will appear here after you run:")
				fmt.Println("   • anstruct aistruct")
				fmt.Println("   • anstruct mstruct")
				fmt.Println("   • anstruct rstruct")
				return nil
			}

			fmt.Printf("📜 Operation History (%d entries):\n\n", len(ops))

			for i, op := range ops {
				// Icon based on type
				icon := "📦"
				switch op.Type {
				case core.OpAI:
					icon = "🤖"
				case core.OpAIApply:
					icon = "🚀"
				case core.OpReverse:
					icon = "🔄"
				case core.OpCreate:
					icon = "📦"
				}

				fmt.Printf("%d. %s [%s] %s\n", i+1, icon, op.Type, op.Timestamp)
				fmt.Printf("   📍 Target: %s\n", op.Target)

				if op.BlueprintPath != "" {
					fmt.Printf("   📝 Blueprint: %s\n", op.BlueprintPath)
				}
				if op.SourcePrompt != "" {
					promptPreview := op.SourcePrompt
					if len(promptPreview) > 50 {
						promptPreview = promptPreview[:50] + "..."
					}
					fmt.Printf("   💬 Prompt: %s\n", promptPreview)
				}

				if len(op.Receipt.CreatedDirs) > 0 || len(op.Receipt.CreatedFiles) > 0 {
					fmt.Printf("   📊 Created: %d dirs, %d files\n",
						len(op.Receipt.CreatedDirs),
						len(op.Receipt.CreatedFiles))
				}
				fmt.Println()
			}

			fmt.Println("💡 Commands:")
			fmt.Println("   anstruct history undo        - Undo last operation")
			fmt.Println("   anstruct history redo        - Redo last undone operation")
			fmt.Println("   anstruct history list --undo-stack - Show redo queue")
			fmt.Println("   anstruct history clear       - Clear all history")
			return nil
		},
	}

	cmd.Flags().BoolVar(&showUndoStack, "undo-stack", false, "show operations that can be redone")
	return cmd
}

func displayUndoStack(histImpl *history.History) error {
	ops, err := histImpl.ListUndoStack(ctx)
	if err != nil {
		return fmt.Errorf("failed to read undo stack: %w", err)
	}

	if len(ops) == 0 {
		fmt.Println("🔄 No operations available for redo")
		fmt.Println("\n💡 Undo an operation first to see it here:")
		fmt.Println("   anstruct history undo --confirm")
		return nil
	}

	fmt.Printf("🔄 Redo Queue (%d entries):\n\n", len(ops))

	for i, op := range ops {
		icon := "📦"
		switch op.Type {
		case core.OpAI:
			icon = "🤖"
		case core.OpAIApply:
			icon = "🚀"
		case core.OpReverse:
			icon = "🔄"
		case core.OpCreate:
			icon = "📦"
		}

		fmt.Printf("%d. %s [%s] %s\n", i+1, icon, op.Type, op.Timestamp)
		fmt.Printf("   📍 Target: %s\n", op.Target)

		if op.BlueprintPath != "" {
			fmt.Printf("   📝 Blueprint: %s\n", op.BlueprintPath)
		}
		if op.SourcePrompt != "" {
			promptPreview := op.SourcePrompt
			if len(promptPreview) > 50 {
				promptPreview = promptPreview[:50] + "..."
			}
			fmt.Printf("   💬 Prompt: %s\n", promptPreview)
		}
		fmt.Println()
	}

	fmt.Println("💡 Use 'anstruct history redo' to restore the last undone operation")
	return nil
}

// newHistoryUndoCmd undoes last operation
func newHistoryUndoCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo the last operation",
		Long: `Undo the last anstruct operation and move it to redo stack.

This will:
- Delete files/folders created by the last operation
- Move operation to redo stack (can be redone)
- Update history

Examples:
  anstruct history undo              # Show warning
  anstruct history undo --confirm    # Execute undo`,

		RunE: func(cmd *cobra.Command, args []string) error {
			histImpl, ok := svc.History.(*history.History)
			if !ok {
				return fmt.Errorf("history implementation error")
			}

			if !confirm {
				fmt.Println("⚠️  WARNING: This will delete files created by the last operation!")
				fmt.Println()

				// Show what will be undone
				ops, err := histImpl.List(ctx)
				if err == nil && len(ops) > 0 {
					last := ops[len(ops)-1]

					icon := "📦"
					switch last.Type {
					case core.OpAI:
						icon = "🤖"
					case core.OpAIApply:
						icon = "🚀"
					case core.OpReverse:
						icon = "🔄"
					}

					fmt.Printf("📍 Will undo: %s [%s] %s\n", icon, last.Type, last.Target)

					if len(last.Receipt.CreatedFiles) > 0 {
						fmt.Printf("   ❌ Will delete: %d files\n", len(last.Receipt.CreatedFiles))
					}
					if len(last.Receipt.CreatedDirs) > 0 {
						fmt.Printf("   ❌ Will delete: %d directories\n", len(last.Receipt.CreatedDirs))
					}

					if last.BlueprintPath != "" {
						fmt.Printf("   📝 Can be recreated from: %s\n", last.BlueprintPath)
					}
					if last.SourcePrompt != "" {
						fmt.Printf("   💬 Can be recreated from AI prompt\n")
					}
				}

				fmt.Println()
				fmt.Println("💡 Use --confirm flag to proceed:")
				fmt.Println("   anstruct history undo --confirm")
				return nil
			}

			if err := histImpl.Undo(ctx); err != nil {
				if err == core.ErrHistoryEmpty {
					fmt.Println("📜 No history to undo")
					return nil
				}
				return fmt.Errorf("undo failed: %w", err)
			}

			fmt.Println("✅ Last operation undone successfully!")
			fmt.Println("💡 Use 'anstruct history redo' to reapply this operation")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&confirm, "confirm", "y", false, "confirm undo without prompt")
	return cmd
}

// newHistoryRedoCmd redoes last undone operation
func newHistoryRedoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redo",
		Short: "Redo the last undone operation",
		Long: `Redo the last operation that was undone.

This will:
- Recreate files/folders from the original blueprint or AI prompt
- Restore operation to history
- Remove from undo stack

Note: Redo will attempt to recreate the exact same structure.

Examples:
  anstruct history redo`,

		RunE: func(cmd *cobra.Command, args []string) error {
			histImpl, ok := svc.History.(*history.History)
			if !ok {
				return fmt.Errorf("history implementation error")
			}

			// Show preview of what will be redone
			ops, err := histImpl.ListUndoStack(ctx)
			if err == nil && len(ops) > 0 {
				last := ops[len(ops)-1]

				icon := "📦"
				switch last.Type {
				case core.OpAI:
					icon = "🤖"
				case core.OpAIApply:
					icon = "🚀"
				case core.OpReverse:
					icon = "🔄"
				}

				fmt.Printf("🔄 Redoing: %s [%s] %s\n", icon, last.Type, last.Target)
				fmt.Println()
			}

			if err := histImpl.Redo(ctx); err != nil {
				if err == core.ErrHistoryEmpty {
					fmt.Println("🔄 No operations to redo")
					fmt.Println("\n💡 Undo an operation first:")
					fmt.Println("   anstruct history undo --confirm")
					return nil
				}
				return fmt.Errorf("redo failed: %w", err)
			}

			fmt.Println("\n✅ Operation redone successfully!")
			fmt.Println("💡 Files have been recreated and operation restored to history")
			return nil
		},
	}

	return cmd
}

// newHistoryClearCmd clears all history
func newHistoryClearCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all history",
		Long: `Clear all operation history and redo stack.

This will:
- Remove all history records
- Clear undo/redo stack
- Cannot be undone

Note: This does NOT delete any actual project files, only the history logs.

Examples:
  anstruct history clear
  anstruct history clear --confirm`,

		RunE: func(cmd *cobra.Command, args []string) error {
			histImpl, ok := svc.History.(*history.History)
			if !ok {
				return fmt.Errorf("history implementation error")
			}

			if !confirm {
				ops, err := histImpl.List(ctx)
				historyCount := 0
				if err == nil {
					historyCount = len(ops)
				}

				undoOps, err := histImpl.ListUndoStack(ctx)
				undoCount := 0
				if err == nil {
					undoCount = len(undoOps)
				}

				fmt.Printf("⚠️  WARNING: This will clear all history!\n\n")
				fmt.Printf("   📜 History entries: %d\n", historyCount)
				fmt.Printf("   🔄 Undo stack: %d\n", undoCount)
				fmt.Println("\n⚠️  This action cannot be undone.")
				fmt.Println("💡 Note: Your actual project files will NOT be deleted.")
				fmt.Println()
				fmt.Println("💡 Use --confirm flag to proceed:")
				fmt.Println("   anstruct history clear --confirm")
				return nil
			}

			if err := histImpl.Clear(ctx); err != nil {
				return fmt.Errorf("failed to clear history: %w", err)
			}

			fmt.Println("✅ History cleared successfully!")
			fmt.Println("💡 Fresh start - all history logs removed")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&confirm, "confirm", "y", false, "confirm clear without prompt")
	return cmd
}
