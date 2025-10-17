## Commands Reference

### `aistruct` - AI-Powered Generation

Generate project structures from natural language.

```bash
anstruct aistruct <prompt> [flags]
```

**Flags:**
- `-o, --out <path>` - Output path (.struct file or folder)
- `--apply` - Generate folder directly (skip blueprint)
- `--dry` - Preview without writing files
- `-v, --verbose` - Show raw AI output
- `--retries <n>` - Retry count if AI output invalid (default: 2)
- `--force` - Overwrite existing files
- `--allow-reserved` - Allow reserved folders (vendor/, node_modules/)

**Examples:**

```bash
# Generate blueprint
anstruct aistruct "golang microservice with gRPC" -o service.struct

# Generate to folder
anstruct aistruct "python flask api" --apply -o ./flask-api

# Output to directory
anstruct aistruct "vue3 app" -o ./blueprints/  # → blueprints/aistruct.struct

# Preview with verbose
anstruct aistruct "rust CLI tool" --dry --verbose

# Allow reserved folders
anstruct aistruct "php laravel api" --apply --allow-reserved -o ./laravel
```

---

### `mstruct` - Generate from Blueprint

Create project from `.struct` blueprint file.

```bash
anstruct mstruct <file.struct> [flags]
```

**Flags:**
- `-o, --out <dir>` - Output directory (default: current folder)
- `--dry` - Simulate without writing
- `--force` - Overwrite existing files
- `-v, --verbose` - Show detailed preview
- `--allow-reserved` - Allow reserved folders

**Examples:**

```bash
# Generate to current directory
anstruct mstruct myapp.struct

# Generate to specific directory
anstruct mstruct myapp.struct -o ./output

# Preview before generation
anstruct mstruct myapp.struct --dry --verbose

# Force overwrite
anstruct mstruct myapp.struct --force -o ./existing-project
```

---

### `rstruct` - Reverse Engineer

Convert existing project to `.struct` blueprint.

```bash
anstruct rstruct <projectDir> [flags]
```

**Flags:**
- `-o, --out <path>` - Output .struct file (auto-detects directory vs file)
- `--dry` - Preview structure without writing
- `-v, --verbose` - Show detailed directory tree

**Examples:**

```bash
# Auto-named output
anstruct rstruct ./myapp  # → myapp.struct

# Custom output file
anstruct rstruct ./myapp -o project.struct

# Output to directory
anstruct rstruct ./myapp -o ./blueprints/  # → blueprints/myapp.struct

# Preview structure
anstruct rstruct ./myapp --dry --verbose
```

---

### `convert` - Normalize Formats

Convert various structure formats to `.struct`.

```bash
anstruct convert <input-file> [flags]
```

**Supported Formats:**
- `tree` - Tree command output (├──, └──, │)
- `ls` - ls -R output
- `markdown` - Markdown with tree symbols
- `plain` - Plain indented text
- `auto` - Auto-detect format (default)

**Normalization Modes:**
- `auto` - Try AI first, fallback to manual (default)
- `ai` - AI-powered normalization only
- `manual` / `offline` - Regex-based parsing (no AI)

**Flags:**
- `-o, --out <file>` - Output .struct file (default: converted.struct)
- `--format <type>` - Input format (auto/tree/ls/markdown/plain)
- `--mode <mode>` - Normalization mode (auto/ai/manual/offline)
- `--stdin` - Read from stdin
- `-v, --verbose` - Show detailed conversion info

**Examples:**

```bash
# From tree command
tree myproject > structure.txt
anstruct convert structure.txt -o myproject.struct

# From stdin
tree myproject | anstruct convert --stdin -o myproject.struct

# With AI normalization
anstruct convert messy-structure.txt --mode ai -o clean.struct

# Offline mode (no AI)
anstruct convert simple-tree.txt --mode offline -o output.struct

# Auto mode with verbose
anstruct convert random-format.txt --mode auto --verbose
```

---

### `watch` - Real-time Sync

Watch and sync project ↔ blueprint in real-time.

```bash
anstruct watch <projectDir> <blueprintFile> [flags]
```

**Modes:**
- `--half struct` - Sync blueprint → folder only
- `--half folder` - Sync folder → blueprint only
- `--full` - Two-way bidirectional sync

**Flags:**
- `--dry` - Simulate without writing
- `-v, --verbose` - Show detailed changes
- `--ignore <pattern>` - Skip files/dirs matching pattern
- `--debounce <duration>` - Delay before reacting (default: 2s)

**Examples:**

```bash
# Two-way sync
anstruct watch ./myapp ./myapp.struct --full

# Blueprint → folder only
anstruct watch ./myapp ./myapp.struct --half struct --verbose

# Folder → blueprint only
anstruct watch ./myapp ./myapp.struct --half folder

# With ignore pattern
anstruct watch ./myapp ./myapp.struct --full --ignore node_modules

# Custom debounce
anstruct watch ./myapp ./myapp.struct --full --debounce 1s
```

---

### `history` - Operation History

Manage operation history with undo/redo support.

```bash
anstruct history <subcommand> [flags]
```

**Subcommands:**
- `list` - Show all operations
- `undo` - Undo last operation
- `redo` - Redo last undone operation
- `clear` - Clear all history

**Examples:**

```bash
# Show history
anstruct history list

# Show redo queue
anstruct history list --undo-stack

# Undo last operation
anstruct history undo --confirm

# Redo last undone
anstruct history redo

# Clear all history
anstruct history clear --confirm
```

---

## .struct Format Specification

The `.struct` format is a simple, human-readable format for defining project structures.

### Syntax Rules

1. **Root folder** - Single folder at top level ending with `/`
2. **Indentation** - Use tabs (`\t`) for hierarchy (2 spaces also supported)
3. **Folders** - End with `/` (e.g., `src/`, `config/`)
4. **Files** - No trailing slash (e.g., `main.go`, `Dockerfile`)
5. **Comments** - Lines starting with `#` are ignored
6. **Empty lines** - Ignored for readability

### Example

```
myapp/
	# Source code
	src/
		main.go
		utils/
			helper.go
			logger.go
	
	# Configuration
	config/
		app.yaml
		database.json
	
	# Docker setup
	Dockerfile
	docker-compose.yml
	
	# Documentation
	README.md
	LICENSE
```

### File vs Folder Detection

| Example | Type | Rule |
|---------|------|------|
| `src/` | Folder | Ends with `/` |
| `main.go` | File | Has extension |
| `Dockerfile` | File | No extension, no `/` |
| `config/` | Folder | Ends with `/` |
| `README.md` | File | Has extension |

---

## Reserved Folders

Anstruct automatically **skips** the following folders as they are auto-generated:

| Folder | Description | Managed By |
|--------|-------------|------------|
| `node_modules/` | NPM/Yarn packages | npm, yarn |
| `vendor/` | Dependencies | Composer, Go modules |
| `.git/` | Git repository | Git |
| `dist/`, `build/` | Build outputs | Build tools |
| `.next/`, `.nuxt/` | Framework builds | Next.js, Nuxt.js |
| `__pycache__/` | Python cache | Python |
| `.venv/`, `venv/` | Virtual environments | Python |
| `.cache/` | Cache directories | Various tools |

**Why skip these?**
- They're auto-generated by package managers
- They're large and change frequently
- They shouldn't be in version control
- They're regenerated from lock files

**Override:** Use `--allow-reserved` flag to include them (not recommended)

```bash
# Skip reserved folders (default)
anstruct aistruct "php laravel api" --apply -o ./api

# Include reserved folders
anstruct aistruct "php laravel api" --apply --allow-reserved -o ./api
```

---

## Use Cases

### 1. Rapid Prototyping

```bash
# Generate complete project structure in seconds
anstruct aistruct "golang REST API with PostgreSQL, Redis, and Docker" --apply -o ./myapi

# Modify structure
vim myapi.struct

# Regenerate
anstruct mstruct myapi.struct -o ./myapi --force
```

### 2. Documentation

```bash
# Document existing project
anstruct rstruct ./complex-project -o docs/structure.struct

# Add to Git
git add docs/structure.struct
```

### 3. Team Onboarding

```bash
# Share blueprint in repository
cat > project.struct << 'EOF'
myproject/
	src/
		api/
		models/
		services/
	tests/
	config/
	README.md
EOF

# New team member clones and generates
git clone <repo>
anstruct mstruct project.struct
```

### 4. Project Migration

```bash
# Convert old structure
anstruct rstruct ./old-project -o old.struct

# Modify for new framework
vim old.struct

# Generate new project
anstruct mstruct old.struct -o ./new-project
```

### 5. Multi-environment Setup

```bash
# Development blueprint
anstruct aistruct "nodejs api for dev" -o dev.struct

# Production blueprint (add optimizations)
cp dev.struct prod.struct
vim prod.struct  # Add caching, monitoring, etc.

# Generate both
anstruct mstruct dev.struct -o ./dev-env
anstruct mstruct prod.struct -o ./prod-env
```

### 6. Live Development Sync

```bash
# Start watch mode
anstruct watch ./myapp ./myapp.struct --full --verbose

# In another terminal, modify files
mkdir ./myapp/new-feature
touch ./myapp/new-feature/handler.go

# Blueprint auto-updates!
cat ./myapp.struct  # Shows new-feature/
```

---

## Advanced Usage

### Custom AI Prompts

For best results, be specific in your prompts:

```bash
# Too vague
anstruct aistruct "web app"

# Specific and detailed
anstruct aistruct "Next.js 14 app with App Router, TypeScript, Tailwind, and tRPC API"

# Include tech stack
anstruct aistruct "Python FastAPI microservice with PostgreSQL, Redis cache, Docker, and pytest"

# Mention structure style
anstruct aistruct "Clean Architecture Golang API with DDD layers: domain, application, infrastructure"
```

### Handling Large Projects

```bash
# Use ignore pattern for watch
anstruct watch ./large-app ./app.struct --full --ignore "node_modules|dist|.cache"

# Generate in stages
anstruct mstruct backend.struct -o ./project/backend
anstruct mstruct frontend.struct -o ./project/frontend
```

### Combining with Other Tools

```bash
# Generate structure, then initialize
anstruct aistruct "golang module" --apply -o ./mymod
cd mymod
go mod init github.com/user/mymod

# Reverse engineer for documentation
anstruct rstruct . -o structure.struct
tree -I 'node_modules|vendor' > structure.txt
anstruct convert structure.txt -o clean.struct

# Sync with Git hooks
# .git/hooks/post-checkout
#!/bin/bash
anstruct mstruct project.struct --force
```

### CI/CD Integration

```yaml
# .github/workflows/validate-structure.yml
name: Validate Structure

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Anstruct
        run: go install github.com/alberdjuniawan/anstruct/cmd/anstruct@latest
      
      - name: Generate from blueprint
        run: anstruct mstruct project.struct -o ./generated --dry
      
      - name: Reverse engineer actual
        run: anstruct rstruct . -o actual.struct
      
      - name: Compare structures
        run: diff project.struct actual.struct
```

---

## Best Practices

### 1. Blueprint Management

```bash
# DO: Version control blueprints
git add *.struct
git commit -m "Add project structure blueprint"

# DO: Use descriptive names
myapp-v1.struct
myapp-microservices.struct
myapp-monolith.struct

# DON'T: Include generated files in blueprints
# vendor/, node_modules/, dist/ are auto-skipped
```

### 2. Folder Naming

```
# DO: Clear, conventional names
src/
tests/
config/
docs/

# DO: Use plural for collections
models/
controllers/
routes/

# AVOID: Unclear abbreviations
misc/
stuff/
tmp/
```

### 3. File Organization

```
myapp/
	# Group by feature (recommended for large apps)
	features/
		auth/
			handler.go
			service.go
			model.go
		products/
			handler.go
			service.go
			model.go
	
	# Or group by type (recommended for small apps)
	handlers/
		auth.go
		products.go
	services/
		auth.go
		products.go
	models/
		user.go
		product.go
```

### 4. Watch Mode Tips

```bash
# DO: Use --half for one-way sync during development
anstruct watch ./myapp ./myapp.struct --half folder

# DO: Use --ignore for noisy directories
anstruct watch ./myapp ./myapp.struct --full --ignore "node_modules|.git|dist"

# DO: Increase debounce for slower systems
anstruct watch ./myapp ./myapp.struct --full --debounce 5s

# ⚠️ CAUTION: --full mode can delete files not in blueprint
# Always commit changes before using --full
```

### 5. AI Generation Tips

```bash
# DO: Review before --apply
anstruct aistruct "your prompt" --dry --verbose
# Review output, then:
anstruct aistruct "your prompt" --apply -o ./project

# DO: Use --retries for better results
anstruct aistruct "complex structure" --retries 3

# DO: Save to blueprint first, review, then generate
anstruct aistruct "your prompt" -o project.struct
vim project.struct  # Review and edit
anstruct mstruct project.struct -o ./project
```

---

## Configuration

### AI Endpoint

```bash
# By default, Anstruct uses the official AI proxy hosted by Alberd for Gemini integration:
https://anstruct-ai-proxy.anstruct.workers.dev/generate
# You don’t need to configure anything — it works out of the box.

# Optionally, you can override the endpoint via environment variable:
export ANSTRUCT_AI_ENDPOINT="https://your-custom-endpoint.com/generate"
```

If you prefer to use your own AI proxy, see the setup guide [here](https://github.com/alberdjuniawan/anstruct-ai-proxy)

### Environment Variables

```bash
# History file location
export ANSTRUCT_HISTORY_PATH="$HOME/.anstruct/history.log"
```

### Directory Structure

```
~/.anstruct/
├── history.log        # Operation history
└── undo_stack.log     # Redo queue
```

---

## Troubleshooting

### AI Generation Issues

**Problem:** AI generates invalid structure

```bash
# Solution 1: Increase retries
anstruct aistruct "your prompt" --retries 5

# Solution 2: Use --verbose to see raw output
anstruct aistruct "your prompt" --verbose

# Solution 3: Be more specific in prompt
anstruct aistruct "Next.js 14 app with src/ directory, app router, and TypeScript"
```

**Problem:** Reserved folders error

```bash
# Solution: Use --allow-reserved (not recommended)
anstruct aistruct "php app" --apply --allow-reserved -o ./phpapp

# Better: Let AI regenerate without reserved folders
anstruct aistruct "php app without vendor directory" --apply -o ./phpapp
```

### Parsing Issues

**Problem:** Indentation errors

```bash
# Solution: Use tabs, not spaces
# Fix manually or use normalize
anstruct convert messy.struct --mode ai -o clean.struct
```

**Problem:** Path traversal detected

```bash
# Cause: Using .. or absolute paths
# BAD:
../outside/
/absolute/path/

# GOOD:
relative/path/
normal-folder/
```

### Watch Mode Issues

**Problem:** Changes not detected

```bash
# Solution 1: Check file permissions
ls -la ./myapp ./myapp.struct

# Solution 2: Increase debounce
anstruct watch ./myapp ./myapp.struct --full --debounce 3s

# Solution 3: Check ignore pattern
anstruct watch ./myapp ./myapp.struct --full --verbose
```

**Problem:** Files deleted unexpectedly

```bash
# Cause: Using --full mode removes files not in blueprint
# Solution: Use --half for one-way sync
anstruct watch ./myapp ./myapp.struct --half struct  # Safe

# Or commit changes first
git add -A && git commit -m "backup before watch"
anstruct watch ./myapp ./myapp.struct --full
```

---

## Performance

### Benchmarks

| Operation | Small Project (10 files) | Medium Project (100 files) | Large Project (1000 files) |
|-----------|-------------------------|----------------------------|----------------------------|
| `mstruct` | < 100ms | < 500ms | < 2s |
| `rstruct` | < 50ms | < 200ms | < 1s |
| `aistruct` | 2-5s (AI dependent) | 2-5s | 2-5s |
| `convert` | < 100ms | < 300ms | < 1s |
| `watch` | Real-time (2s debounce) | Real-time | Real-time |

### Optimization Tips

```bash
# Use --dry to preview before generation
anstruct mstruct large.struct --dry  # Fast preview

# Generate incrementally
anstruct mstruct backend.struct -o ./project/backend
anstruct mstruct frontend.struct -o ./project/frontend

# Use .gitignore patterns for watch
anstruct watch ./app ./app.struct --full --ignore ".git|node_modules|dist|*.log"
```

---
