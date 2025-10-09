package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alberdjuniawan/anstruct"
	"github.com/alberdjuniawan/anstruct/internal/core"
	"github.com/spf13/cobra"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: anstruct <command> [options]")
		fmt.Println("Commands: aistruct, mstruct, rstruct, watch")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	ctx := context.Background()
	endpoint := "https://anstruct-ai-proxy.anstruct.workers.dev"
	historyPath := ".anstruct/history.log"
	svc := anstruct.NewService(endpoint, historyPath)
	rootCmd := &cobra.Command{Use: "anstruct"}
	rootCmd.AddCommand(newWatchCmd(svc))

	switch cmd {

	case "aistruct":
		fs := flag.NewFlagSet("aistruct", flag.ExitOnError)
		outFile := fs.String("out", "aistruct.struct", "output .struct file")
		fs.Parse(args)
		if fs.NArg() == 0 {
			fmt.Println("Usage: anstruct aistruct [options] <prompt>")
			os.Exit(1)
		}
		prompt := fs.Arg(0)
		if err := svc.AIStruct(ctx, prompt, *outFile); err != nil {
			fmt.Println("AIStruct error:", err)
			os.Exit(1)
		}
		fmt.Println("Blueprint written to", *outFile)

	case "mstruct":
		fs := flag.NewFlagSet("mstruct", flag.ExitOnError)
		outDir := fs.String("out", ".", "output directory")

		var dry bool
		var force bool
		fs.BoolVar(&dry, "dry", false, "dry run")
		fs.BoolVar(&force, "force", false, "overwrite existing files")

		fs.Parse(args)
		if fs.NArg() == 0 {
			fmt.Println("Usage: anstruct mstruct [options] <file.struct>")
			os.Exit(1)
		}
		structFile := fs.Arg(0)
		receipt, err := svc.MStruct(ctx, structFile, *outDir, core.GenerateOptions{DryRun: dry, Force: force})
		if err != nil {
			fmt.Println("MStruct error:", err)
			os.Exit(1)
		}
		fmt.Println("Generated:", receipt.CreatedDirs, receipt.CreatedFiles)

	case "rstruct":
		fs := flag.NewFlagSet("rstruct", flag.ExitOnError)
		outFile := fs.String("out", "", "output .struct file (default: ./<folder>.struct)")
		fs.Parse(args)
		if fs.NArg() == 0 {
			fmt.Println("Usage: anstruct rstruct [options] <folder>")
			os.Exit(1)
		}
		folder := fs.Arg(0)

		// Tentukan path output sesuai rules
		if *outFile == "" {
			base := filepath.Base(folder)
			*outFile = filepath.Join(".", base+".struct")
		} else {
			cleaned := filepath.Clean(*outFile)
			if strings.HasSuffix(cleaned, ".struct") {
				*outFile = cleaned
			} else {
				base := filepath.Base(folder)
				// bedanya di sini 👇
				*outFile = filepath.Join(cleaned, base) + ".struct"
			}
		}

		fmt.Println("DEBUG main.go resolved outFile =", *outFile)

		if err := svc.RStruct(ctx, folder, *outFile); err != nil {
			fmt.Println("RStruct error:", err)
			os.Exit(1)
		}

		fmt.Println("Blueprint written to", *outFile)

		absOut, _ := filepath.Abs(*outFile)
		absFolder, _ := filepath.Abs(folder)
		if filepath.Dir(absOut) == absFolder {
			fmt.Println("Note: .struct file is inside the project folder. It will not be listed in the blueprint itself, and watcher will ignore it.")
		} else {
			fmt.Println("Tip: use -out <folder> if you want the blueprint inside the project folder, or -out <file.struct> for a custom name.")
		}

	case "watch":
		fs := flag.NewFlagSet("watch", flag.ExitOnError)
		verbose := fs.Bool("v", false, "verbose logging")
		debounce := fs.Duration("debounce", 2*time.Second, "debounce interval")
		fs.Parse(args)

		if fs.NArg() < 2 {
			fmt.Println("Usage: anstruct watch [options] <projectDir> <blueprintFile>")
			os.Exit(1)
		}
		project := fs.Arg(0)
		blueprint := fs.Arg(1)

		fmt.Println("Watching project:", project, "and blueprint:", blueprint)
		if err := svc.Watch(ctx, project, blueprint, *debounce, *verbose); err != nil {
			fmt.Println("Watch error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown command:", cmd)
		fmt.Println("Commands: aistruct, mstruct, rstruct, watch")
		os.Exit(1)
	}
}
