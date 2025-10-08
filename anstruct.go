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

// Facade: expose fungsi publik
type Service struct {
	Gen       core.Generator
	Parser    core.Parser
	Reverser  core.Reverser
	Validator core.Validator
	History   core.History
	Writer    *generator.Generator
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

func (s *Service) GenerateFromPrompt(ctx context.Context, prompt, outputDir string, opts core.GenerateOptions) (core.Receipt, error) {
	tree, err := s.Gen.FromPrompt(ctx, prompt)
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

func (s *Service) ReverseProject(ctx context.Context, inputDir, blueprintPath string) error {
	tree, err := s.Reverser.Reverse(ctx, inputDir)
	if err != nil {
		return err
	}
	if err := s.Parser.Write(ctx, tree, blueprintPath); err != nil {
		return err
	}
	_ = s.History.Record(ctx, core.Operation{Type: core.OpReverse, Target: blueprintPath})
	return nil
}

// Watch: sinkronisasi folder <-> blueprint
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
			// folder → blueprint
			if err := s.ReverseProject(ctx, projectPath, blueprintPath); err != nil && verbose {
				fmt.Println("Reverse error:", err)
			}
		},
		func() {
			// blueprint → folder
			tree, err := s.Parser.Parse(ctx, blueprintPath)
			if err == nil {
				_, _ = s.Writer.Generate(ctx, tree, projectPath, core.GenerateOptions{Force: true})
			} else if verbose {
				fmt.Println("Parse error:", err)
			}
		},
	)
}
