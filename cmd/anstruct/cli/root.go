package cli

import (
	"context"
	"os"

	"github.com/alberdjuniawan/anstruct"
	"github.com/spf13/cobra"
)

var (
	svc      *anstruct.Service
	rootCmd  *cobra.Command
	ctx      context.Context
	endpoint = "https://anstruct-ai-proxy.anstruct.workers.dev"
	history  = ".anstruct/history.log"
)

func init() {
	ctx = context.Background()
	svc = anstruct.NewService(endpoint, history)

	rootCmd = &cobra.Command{
		Use:   "anstruct",
		Short: "Anstruct - AI-powered project structure manager",
	}

	rootCmd.AddCommand(
		newAIStructCmd(),
		newMStructCmd(),
		newRStructCmd(),
		newWatchCmd(svc),
		newUndoCmd(),
		newHistoryCmd(),
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
