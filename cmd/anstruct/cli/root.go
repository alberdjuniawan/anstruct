package cli

import (
	"context"
	"os"

	"github.com/alberdjuniawan/anstruct"
	"github.com/spf13/cobra"
)

var (
	svc         *anstruct.Service
	rootCmd     *cobra.Command
	ctx         context.Context
	endpoint    = "https://anstruct-ai-proxy.anstruct.workers.dev"
	historyPath = ".anstruct/history.log"
)

func init() {
	ctx = context.Background()
	svc = anstruct.NewService(endpoint, historyPath)

	rootCmd = &cobra.Command{
		Use:   "anstruct",
		Short: "Anstruct - AI-powered project structure manager",
		Long: `Anstruct is a powerful CLI tool for managing project structures using AI.

Core Commands:
  aistruct   - Generate structure from natural language
  mstruct    - Create project from .struct blueprint
  rstruct    - Reverse engineer project to blueprint
  normalize  - Convert any format to .struct format
  
Utility Commands:
  watch      - Watch and sync project ↔ blueprint
  history    - Manage operation history (undo/redo)

Examples:
  anstruct aistruct "nodejs api with auth" --apply -o ./myapi
  anstruct mstruct myproject.struct -o ./output
  anstruct rstruct ./myapp -o myapp.struct
  anstruct normalize structure.txt -o project.struct
  anstruct history undo --confirm`,
	}

	rootCmd.AddCommand(
		newAIStructCmd(),
		newMStructCmd(),
		newRStructCmd(),
		newNormalizeCmd(), // NEW: Normalize command
		newWatchCmd(svc),
		newHistoryCmd(),
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
