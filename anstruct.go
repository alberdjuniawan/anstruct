package anstruct

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alberdjuniawan/anstruct/internal/ai"
	"github.com/alberdjuniawan/anstruct/internal/core"
	"github.com/alberdjuniawan/anstruct/internal/generator"
	"github.com/alberdjuniawan/anstruct/internal/history"
	"github.com/alberdjuniawan/anstruct/internal/parser"
	"github.com/alberdjuniawan/anstruct/internal/reverser"
	"github.com/alberdjuniawan/anstruct/internal/validator"
	"github.com/alberdjuniawan/anstruct/internal/watcher"
)

type Service struct {
	Gen       *ai.AIGenerator
	Parser    core.Parser
	Reverser  core.Reverser
	Validator core.Validator
	History   core.History
	Writer    *generator.Generator
}

func NewService(endpoint, historyPath string) *Service {
	p := parser.New()

	provider := ai.NewGeminiProvider(endpoint)

	return &Service{
		Gen:       ai.NewAIGenerator(provider, p),
		Parser:    p,
		Reverser:  reverser.New(),
		Validator: validator.New(),
		History:   history.New(historyPath),
		Writer:    generator.New(),
	}
}

func (s *Service) AIStruct(ctx context.Context, prompt, outPath string, opts core.AIOptions) error {
	fmt.Printf("🤖 Generating from prompt: %s\n", prompt)

	tree, rawOutput, err := s.Gen.FromPrompt(ctx, prompt, opts.Retries)

	if opts.Verbose && rawOutput != "" {
		fmt.Println("\n📋 Raw AI Output:")
		fmt.Println("---")
		fmt.Println(rawOutput)
		fmt.Println("---")
	}

	if err != nil {
		if rawOutput != "" {
			fallbackFile := "ai_invalid_" + time.Now().Format("20060102_150405") + ".struct"
			_ = os.WriteFile(fallbackFile, []byte(rawOutput), 0644)
			return fmt.Errorf("%w\n💾 Raw output saved to: %s", err, fallbackFile)
		}
		return err
	}

	if opts.DryRun {
		fmt.Println("\n📂 Preview of generated structure:")
		displayTree(tree.Root, 0)
		fmt.Printf("\n✅ Dry run complete. No files written.\n")
		return nil
	}

	if opts.Apply {
		return s.applyDirectly(ctx, tree, outPath, opts)
	}

	return s.saveBlueprint(ctx, tree, outPath)
}

func (s *Service) applyDirectly(ctx context.Context, tree *core.Tree, outPath string, opts core.AIOptions) error {
	fmt.Printf("\n📁 Generating project folder: %s\n", outPath)

	if err := s.Validator.Validate(ctx, tree); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	receipt, err := s.Writer.Generate(ctx, tree, outPath, core.GenerateOptions{
		DryRun: false,
		Force:  opts.Force,
	})
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	_ = s.History.Record(ctx, core.Operation{
		Type:    core.OpAIApply,
		Target:  outPath,
		Receipt: receipt,
	})

	fmt.Printf("\n✅ Project generated successfully!\n")
	fmt.Printf("   📂 %d directories created\n", len(receipt.CreatedDirs))
	fmt.Printf("   📄 %d files created\n", len(receipt.CreatedFiles))
	fmt.Printf("\n📍 Location: %s\n", outPath)

	return nil
}

func (s *Service) saveBlueprint(ctx context.Context, tree *core.Tree, outPath string) error {
	fmt.Printf("\n📝 Saving blueprint: %s\n", outPath)

	dir := filepath.Dir(outPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := s.Parser.Write(ctx, tree, outPath); err != nil {
		return fmt.Errorf("failed to write blueprint: %w", err)
	}

	_ = s.History.Record(ctx, core.Operation{
		Type:   core.OpAI,
		Target: outPath,
	})

	fmt.Printf("✅ Blueprint saved: %s\n", outPath)
	return nil
}

func displayTree(n *core.Node, depth int) {
	if depth > 0 {
		indent := ""
		for i := 0; i < depth-1; i++ {
			indent += "  "
		}
		symbol := "📄"
		if n.Type == core.NodeDir {
			symbol = "📁"
		}
		fmt.Printf("%s%s %s\n", indent, symbol, n.Name)
	}
	for _, c := range n.Children {
		displayTree(c, depth+1)
	}
}

func (s *Service) MStruct(ctx context.Context, structFile, outputDir string, opts core.GenerateOptions) (core.Receipt, error) {
	tree, err := s.Parser.Parse(ctx, structFile)
	if err != nil {
		return core.Receipt{}, err
	}
	if err := s.Validator.Validate(ctx, tree); err != nil {
		return core.Receipt{}, err
	}
	receipt, err := s.Writer.Generate(ctx, tree, outputDir, opts)
	if err != nil {
		return receipt, err
	}
	_ = s.History.Record(ctx, core.Operation{Type: core.OpCreate, Target: outputDir, Receipt: receipt})
	return receipt, nil
}

func (s *Service) RStruct(ctx context.Context, inputDir string, outPath string) error {
	tree, err := s.Reverser.Reverse(ctx, inputDir)
	if err != nil {
		return err
	}
	if err := s.Parser.Write(ctx, tree, outPath); err != nil {
		return err
	}
	_ = s.History.Record(ctx, core.Operation{Type: core.OpReverse, Target: outPath})
	return nil
}

func (s *Service) Watch(ctx context.Context, projectPath, blueprintPath string, debounce time.Duration, verbose bool) error {
	w := watcher.New()
	cfg := watcher.SyncConfig{
		ProjectPath:   projectPath,
		BlueprintPath: blueprintPath,
		Debounce:      debounce,
		Verbose:       verbose,
	}
	return w.Run(ctx, cfg,
		func() {
			if err := s.RStruct(ctx, projectPath, blueprintPath); err != nil && verbose {
				fmt.Println("Reverse error:", err)
			}
		},
		func() {
			tree, err := s.Parser.Parse(ctx, blueprintPath)
			if err == nil {
				receipt, genErr := s.Writer.Generate(ctx, tree, projectPath, core.GenerateOptions{Force: true})
				if genErr == nil {
					allowed := map[string]bool{}
					for _, c := range tree.Root.Children {
						generator.CollectAllowed(c, "", allowed)
					}
					if err := generator.CleanupExtra(projectPath, allowed); err != nil && verbose {
						fmt.Println("Cleanup error:", err)
					}
					if verbose {
						fmt.Println("✅ Synced:", receipt.CreatedDirs, receipt.CreatedFiles)
					}
				} else if verbose {
					fmt.Println("Generate error:", genErr)
				}
			} else if verbose {
				fmt.Println("Parse error:", err)
			}
		},
	)
}
