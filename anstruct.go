package anstruct

import (
	"context"
	"fmt"
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
	Gen       core.Generator // AI generator (FromPrompt)
	Parser    core.Parser
	Reverser  core.Reverser
	Validator core.Validator
	History   core.History
	Writer    *generator.Generator // file/folder generator
}

func NewService(endpoint, historyPath string) *Service {
	p := parser.New()
	return &Service{
		Gen:       ai.NewAIGenerator(ai.NewGeminiProvider(endpoint), p),
		Parser:    p,
		Reverser:  reverser.New(),
		Validator: validator.New(),
		History:   history.New(historyPath),
		Writer:    generator.New(),
	}
}

// aistruct: prompt → .struct file
func (s *Service) AIStruct(ctx context.Context, prompt, outFile string) error {
	tree, err := s.Gen.FromPrompt(ctx, prompt)
	if err != nil {
		return err
	}
	if err := s.Parser.Write(ctx, tree, outFile); err != nil {
		return err
	}
	_ = s.History.Record(ctx, core.Operation{Type: core.OpCreate, Target: outFile})
	return nil
}

// mstruct: .struct file → folder
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

// rstruct: folder → .struct file
func (s *Service) RStruct(ctx context.Context, inputDir string, outPath string) error {
	fmt.Println("DEBUG RStruct outPath =", outPath)
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

// watch: sinkronisasi folder <-> blueprint
func (s *Service) Watch(ctx context.Context, projectPath, blueprintPath string, debounce time.Duration, verbose bool) error {
	w := watcher.New()
	cfg := watcher.SyncConfig{
		ProjectPath:   projectPath,
		BlueprintPath: blueprintPath,
		Debounce:      debounce,
		Verbose:       verbose,
	}
	return w.Run(ctx, cfg,
		// onFolder
		func() {
			if err := s.RStruct(ctx, projectPath, blueprintPath); err != nil && verbose {
				fmt.Println("Reverse error:", err)
			}
		},
		// onBlueprint
		func() {
			tree, err := s.Parser.Parse(ctx, blueprintPath)
			if err == nil {
				receipt, genErr := s.Writer.Generate(ctx, tree, projectPath, core.GenerateOptions{Force: true})
				if genErr == nil {
					// full sync cleanup
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
