package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/alberdjuniawan/anstruct"
	"github.com/alberdjuniawan/anstruct/internal/core"
)

func main() {
	var (
		endpoint     = flag.String("endpoint", "https://anstruct-ai-proxy.anstruct.workers.dev", "AI proxy endpoint")
		historyPath  = flag.String("history", ".anstruct/history.log", "history log path")
		outputDir    = flag.String("out", ".", "output directory")
		reverseDir   = flag.String("reverse", "", "reverse project dir")
		blueprintOut = flag.String("blueprint", "reversed.struct", "output blueprint path")
		dryRun       = flag.Bool("dry", false, "dry run (no files created)")
		force        = flag.Bool("force", false, "overwrite existing files")
		watch        = flag.Bool("watch", false, "watch project and blueprint for changes")
		verbose      = flag.Bool("v", false, "verbose logging")
	)
	flag.Parse()

	ctx := context.Background()
	svc := anstruct.NewService(*endpoint, *historyPath)

	// Reverse mode
	if *reverseDir != "" {
		if err := svc.ReverseProject(ctx, *reverseDir, *blueprintOut); err != nil {
			fmt.Println("Reverse error:", err)
			os.Exit(1)
		}
		fmt.Println("Blueprint written to", *blueprintOut)
		return
	}

	// Watch mode
	if *watch {
		fmt.Println("Watching project:", *outputDir, "and blueprint:", *blueprintOut)
		if err := svc.Watch(ctx, *outputDir, *blueprintOut, 2*time.Second, *verbose); err != nil {
			fmt.Println("Watch error:", err)
			os.Exit(1)
		}
		return
	}

	// Generate mode
	if flag.NArg() == 0 {
		fmt.Println("Usage: anstruct [options] <prompt>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	prompt := flag.Arg(0)
	receipt, err := svc.GenerateFromPrompt(ctx, prompt, *outputDir, core.GenerateOptions{DryRun: *dryRun, Force: *force})
	if err != nil {
		fmt.Println("Generate error:", err)
		os.Exit(1)
	}
	fmt.Println("Generated:", receipt.CreatedDirs, receipt.CreatedFiles)
}
