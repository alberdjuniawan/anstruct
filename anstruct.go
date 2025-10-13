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
	"github.com/alberdjuniawan/anstruct/internal/normalizer"
	"github.com/alberdjuniawan/anstruct/internal/parser"
	"github.com/alberdjuniawan/anstruct/internal/reverser"
	"github.com/alberdjuniawan/anstruct/internal/validator"
	"github.com/alberdjuniawan/anstruct/internal/watcher"
)

type Service struct {
	Gen        *ai.AIGenerator
	Parser     core.Parser
	Reverser   core.Reverser
	Validator  core.Validator
	History    core.History
	Writer     *generator.Generator
	Normalizer *normalizer.Normalizer
}

// OperationRecreator implements history.OperationRecreator
type OperationRecreator struct {
	svc *Service
}

func NewService(endpoint, historyPath string) *Service {
	p := parser.New()
	provider := ai.NewGeminiProvider(endpoint)

	s := &Service{
		Gen:        ai.NewAIGenerator(provider, p),
		Parser:     p,
		Reverser:   reverser.New(),
		Validator:  validator.New(),
		History:    history.New(historyPath),
		Writer:     generator.New(),
		Normalizer: normalizer.New(),
	}

	// Setup recreator untuk redo functionality
	recreator := &OperationRecreator{svc: s}
	if hist, ok := s.History.(*history.History); ok {
		hist.SetRecreator(recreator)
	}

	return s
}

// RecreateOperation implements redo logic for different operation types
func (r *OperationRecreator) RecreateOperation(ctx context.Context, op core.Operation) error {
	switch op.Type {
	case core.OpCreate:
		return r.recreateCreate(ctx, op)
	case core.OpAIApply:
		return r.recreateAIApply(ctx, op)
	case core.OpReverse:
		return r.recreateReverse(ctx, op)
	case core.OpAI:
		return r.recreateAIBlueprint(ctx, op)
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

func (r *OperationRecreator) recreateCreate(ctx context.Context, op core.Operation) error {
	if op.BlueprintPath == "" {
		return fmt.Errorf("cannot recreate: blueprint path not saved in operation")
	}

	if _, err := os.Stat(op.BlueprintPath); os.IsNotExist(err) {
		return fmt.Errorf("cannot recreate: blueprint file not found: %s", op.BlueprintPath)
	}

	fmt.Printf("🔄 Recreating from blueprint: %s\n", op.BlueprintPath)

	tree, err := r.svc.Parser.Parse(ctx, op.BlueprintPath)
	if err != nil {
		return fmt.Errorf("failed to parse blueprint: %w", err)
	}

	if err := r.svc.Validator.Validate(ctx, tree); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	receipt, err := r.svc.Writer.Generate(ctx, tree, op.Target, core.GenerateOptions{
		DryRun: false,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	op.Receipt = receipt
	fmt.Printf("✅ Recreated: %d dirs, %d files\n", len(receipt.CreatedDirs), len(receipt.CreatedFiles))
	return nil
}

func (r *OperationRecreator) recreateAIApply(ctx context.Context, op core.Operation) error {
	if op.SourcePrompt == "" {
		return fmt.Errorf("cannot recreate: source prompt not saved in operation")
	}

	fmt.Printf("🤖 Regenerating from AI prompt: %s\n", op.SourcePrompt)

	tree, _, err := r.svc.Gen.FromPrompt(ctx, op.SourcePrompt, 2)
	if err != nil {
		return fmt.Errorf("AI generation failed: %w", err)
	}

	if err := r.svc.Validator.Validate(ctx, tree); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	receipt, err := r.svc.Writer.Generate(ctx, tree, op.Target, core.GenerateOptions{
		DryRun: false,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	op.Receipt = receipt
	fmt.Printf("✅ Regenerated: %d dirs, %d files\n", len(receipt.CreatedDirs), len(receipt.CreatedFiles))
	return nil
}

func (r *OperationRecreator) recreateReverse(ctx context.Context, op core.Operation) error {
	return fmt.Errorf("cannot automatically recreate reverse operation - please run 'anstruct rstruct' manually")
}

func (r *OperationRecreator) recreateAIBlueprint(ctx context.Context, op core.Operation) error {
	if op.SourcePrompt == "" {
		return fmt.Errorf("cannot recreate: source prompt not saved in operation")
	}

	fmt.Printf("🤖 Regenerating AI blueprint from prompt: %s\n", op.SourcePrompt)

	tree, rawOutput, err := r.svc.Gen.FromPrompt(ctx, op.SourcePrompt, 2)
	if err != nil {
		if rawOutput != "" {
			fallbackFile := "ai_invalid_" + time.Now().Format("20060102_150405") + ".struct"
			_ = os.WriteFile(fallbackFile, []byte(rawOutput), 0644)
			return fmt.Errorf("%w\n💾 Raw output saved to: %s", err, fallbackFile)
		}
		return err
	}

	dir := filepath.Dir(op.Target)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := r.svc.Parser.Write(ctx, tree, op.Target); err != nil {
		return fmt.Errorf("failed to write blueprint: %w", err)
	}

	fmt.Printf("✅ Blueprint recreated: %s\n", op.Target)
	return nil
}

// AIStruct generates project structure from natural language
func (s *Service) AIStruct(ctx context.Context, prompt, outPath string, opts core.AIOptions) error {
	fmt.Printf("🤖 Generating from prompt: %s\n", prompt)

	// FromPrompt already uses CleanAIOutput internally (in ai/generator.go)
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
		return s.applyDirectly(ctx, tree, outPath, prompt, opts)
	}

	return s.saveBlueprint(ctx, tree, outPath, prompt)
}

func (s *Service) applyDirectly(ctx context.Context, tree *core.Tree, outPath, prompt string, opts core.AIOptions) error {
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
		Type:         core.OpAIApply,
		Target:       outPath,
		Receipt:      receipt,
		SourcePrompt: prompt,
	})

	fmt.Printf("\n✅ Project generated successfully!\n")
	fmt.Printf("   📂 %d directories created\n", len(receipt.CreatedDirs))
	fmt.Printf("   📄 %d files created\n", len(receipt.CreatedFiles))
	fmt.Printf("\n📍 Location: %s\n", outPath)

	return nil
}

func (s *Service) saveBlueprint(ctx context.Context, tree *core.Tree, outPath, prompt string) error {
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
		Type:         core.OpAI,
		Target:       outPath,
		SourcePrompt: prompt,
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

// MStruct generates project from .struct blueprint file
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

	_ = s.History.Record(ctx, core.Operation{
		Type:          core.OpCreate,
		Target:        outputDir,
		Receipt:       receipt,
		BlueprintPath: structFile,
	})

	return receipt, nil
}

// RStruct reverse engineers a project folder into .struct blueprint
func (s *Service) RStruct(ctx context.Context, inputDir string, outPath string) error {
	tree, err := s.Reverser.Reverse(ctx, inputDir)
	if err != nil {
		return err
	}
	if err := s.Parser.Write(ctx, tree, outPath); err != nil {
		return err
	}

	_ = s.History.Record(ctx, core.Operation{
		Type:   core.OpReverse,
		Target: outPath,
	})

	return nil
}

// NormalizeStruct converts any text-based structure format to .struct format
// Strategy: Try rule-based first (fast, offline), fallback to AI if low confidence
func (s *Service) NormalizeStruct(ctx context.Context, inputContent, outPath string, opts core.AIOptions) error {
	fmt.Println("🔄 Normalizing structure format...")

	// Step 1: Try rule-based normalization (offline, fast)
	normalized, confidence, err := s.Normalizer.NormalizeToStruct(inputContent)

	if err == nil && confidence >= 70 {
		// High confidence - use rule-based result
		fmt.Printf("✨ Normalized using pattern matching (confidence: %d%%)\n", confidence)

		if validateErr := s.Normalizer.ValidateStructOutput(normalized); validateErr != nil {
			fmt.Printf("⚠️  Validation failed: %v\n", validateErr)
			fmt.Println("📡 Falling back to AI normalization...")
			return s.normalizeWithAI(ctx, inputContent, outPath, opts)
		}

		return s.saveNormalizedOutput(ctx, normalized, outPath, opts)
	}

	// Step 2: Low confidence or error - use AI normalization
	if confidence < 70 {
		fmt.Printf("⚠️  Low confidence (%d%%), using AI normalization...\n", confidence)
	} else {
		fmt.Printf("⚠️  Rule-based normalization failed: %v\n", err)
		fmt.Println("📡 Falling back to AI normalization...")
	}

	return s.normalizeWithAI(ctx, inputContent, outPath, opts)
}

// normalizeWithAI uses AI to normalize structure (with proper cleaning)
func (s *Service) normalizeWithAI(ctx context.Context, inputContent, outPath string, opts core.AIOptions) error {
	// Build normalize prompt using ai.NormalizePrompt()
	prompt := ai.NormalizePrompt(inputContent)

	if opts.Verbose {
		fmt.Println("\n📋 Normalize Prompt:")
		fmt.Println("---")
		fmt.Println(prompt)
		fmt.Println("---")
	}

	// Generate using AI (this internally uses CleanAIOutput)
	tree, rawOutput, err := s.Gen.FromPrompt(ctx, prompt, opts.Retries)

	if opts.Verbose && rawOutput != "" {
		fmt.Println("\n📋 AI Raw Output:")
		fmt.Println("---")
		fmt.Println(rawOutput)
		fmt.Println("---")
	}

	if err != nil {
		if rawOutput != "" {
			fallbackFile := "normalize_invalid_" + time.Now().Format("20060102_150405") + ".txt"
			_ = os.WriteFile(fallbackFile, []byte(rawOutput), 0644)
			return fmt.Errorf("%w\n💾 Raw output saved to: %s", err, fallbackFile)
		}
		return fmt.Errorf("AI normalization failed: %w", err)
	}

	if opts.DryRun {
		fmt.Println("\n📂 Preview of normalized structure:")
		displayTree(tree.Root, 0)
		fmt.Printf("\n✅ Dry run complete. No files written.\n")
		return nil
	}

	// Save normalized blueprint
	dir := filepath.Dir(outPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := s.Parser.Write(ctx, tree, outPath); err != nil {
		return fmt.Errorf("failed to write normalized blueprint: %w", err)
	}

	fmt.Printf("\n✅ Structure normalized and saved to: %s\n", outPath)
	fmt.Println("💡 You can now use: anstruct mstruct " + outPath)
	return nil
}

// saveNormalizedOutput saves rule-based normalized output
func (s *Service) saveNormalizedOutput(ctx context.Context, normalized, outPath string, opts core.AIOptions) error {
	if opts.Verbose {
		fmt.Println("\n📋 Normalized Output:")
		fmt.Println("---")
		fmt.Println(normalized)
		fmt.Println("---")
	}

	// Parse the normalized content to validate it's proper .struct
	tree, err := s.Parser.ParseString(ctx, normalized)
	if err != nil {
		return fmt.Errorf("failed to parse normalized output: %w", err)
	}

	if opts.DryRun {
		fmt.Println("\n📂 Preview of normalized structure:")
		displayTree(tree.Root, 0)
		fmt.Printf("\n✅ Dry run complete. No files written.\n")
		return nil
	}

	// Save to file
	dir := filepath.Dir(outPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(outPath, []byte(normalized), 0644); err != nil {
		return fmt.Errorf("failed to write normalized file: %w", err)
	}

	fmt.Printf("\n✅ Structure normalized and saved to: %s\n", outPath)
	fmt.Println("💡 You can now use: anstruct mstruct " + outPath)
	return nil
}

// Watch provides real-time sync between project and blueprint
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
						fmt.Println("✅ Synced:", len(receipt.CreatedDirs), "dirs,", len(receipt.CreatedFiles), "files")
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
