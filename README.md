# Anstruct - AI-Powered Project Structure Manager

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**Anstruct** is a powerful CLI tool that revolutionizes project structure management using AI. Generate, reverse-engineer, normalize, and sync project structures with simple commands.

```bash
# Generate a project from natural language
anstruct aistruct "golang REST API with PostgreSQL" --apply -o ./my-api

# Reverse engineer existing project
anstruct rstruct ./my-project -o project.struct

# Watch and sync in real-time
anstruct watch ./my-app ./my-app.struct --full
```

---

## Features

- **AI-Powered Generation** - Create project structures from natural language
- **Blueprint System** - Define structures in simple `.struct` format
- **Reverse Engineering** - Convert existing projects to blueprints
- **Format Normalization** - Convert any structure format (tree, ls, markdown) to `.struct`
- **Real-time Sync** - Watch and sync project ↔ blueprint bidirectionally
- **History Management** - Undo/redo operations with full tracking
- **Fast & Efficient** - Written in Go for maximum performance

---

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/alberdjuniawan/anstruct.git
cd anstruct

# Build
go build -o anstruct ./cmd/anstruct

# Install globally (optional)
go install ./cmd/anstruct
```

### Using Go Install

```bash
go install github.com/alberdjuniawan/anstruct/cmd/anstruct@latest
```

## Quick Start

### 1. Generate from AI Prompt

```bash
# Generate blueprint file
anstruct aistruct "nodejs express api with auth" -o api.struct

# Generate project directly
anstruct aistruct "react dashboard with routing" --apply -o ./dashboard
```

### 2. Create from Blueprint

```bash
# Create myapp.struct
cat > myapp.struct << 'EOF'
myapp/
	src/
		main.go
		routes/
			api.go
	config/
		app.yaml
	Dockerfile
	README.md
EOF

# Generate project
anstruct mstruct myapp.struct -o ./output
```

### 3. Reverse Engineer Project

```bash
# Convert project to blueprint
anstruct rstruct ./my-existing-app -o app.struct
```

### 4. Watch & Sync

```bash
# Two-way sync
anstruct watch ./myapp ./myapp.struct --full

# One-way: blueprint → folder
anstruct watch ./myapp ./myapp.struct --half struct

# One-way: folder → blueprint
anstruct watch ./myapp ./myapp.struct --half folder
```

---

## 🤝 Contributing

We welcome contributions! Here's how you can help:

### Reporting Bugs

```bash
# Include:
# 1. Anstruct version
anstruct --version

# 2. Command that caused issue
anstruct aistruct "your prompt" --verbose

# 3. Expected vs actual behavior
# 4. Operating system and Go version
```

### Feature Requests

Open an issue with:
- Clear use case
- Expected behavior
- Example commands

### Development Setup

```bash
# Clone repository
git clone https://github.com/alberdjuniawan/anstruct.git
cd anstruct

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o anstruct ./cmd/anstruct

# Run locally
./anstruct --help
```

### Code Structure

```
anstruct/
├── cmd/anstruct/          # CLI entry point
│   ├── main.go
│   └── cli/               # Command implementations
├── internal/
│   ├── ai/                # AI generation logic
│   ├── converter/         # Format conversion
│   ├── core/              # Core types and interfaces
│   ├── generator/         # File/folder generation
│   ├── history/           # History management
│   ├── parser/            # .struct parser
│   ├── reverser/          # Reverse engineering
│   ├── validator/         # Structure validation
│   └── watcher/           # File watching
├── anstruct.go            # Main service
└── README.md
```

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- Powered by [Gemini AI](https://deepmind.google/technologies/gemini/)
- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- File watching by [fsnotify](https://github.com/fsnotify/fsnotify)

---

## Support

- **Issues:** [GitHub Issues](https://github.com/alberdjuniawan/anstruct/issues)
- **Discussions:** [GitHub Discussions](https://github.com/alberdjuniawan/anstruct/discussions)
- **Email:** alberdjuniawanpasunda@gmail.com

---

<div align="center">

**Made by [Alberd Juniawan](https://github.com/alberdjuniawan)**

[⭐ Star on GitHub](https://github.com/alberdjuniawan/anstruct) • [Report Bug](https://github.com/alberdjuniawan/anstruct/issues) • [Request Feature](https://github.com/alberdjuniawan/anstruct/issues)

</div>
