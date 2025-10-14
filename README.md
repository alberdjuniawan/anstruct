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

## Documentation

You can find full guides and references in the [Documentation file](https://github.com/alberdjuniawan/anstruct/blob/main/docs/DOCUMENTATION.md). It covers everything from basic usage to advanced features — including CLI commands, structure syntax, and AI integration.

## Contributing

Contributions are welcome. If you'd like to suggest improvements, fix bugs, or enhance features, please see the [Contributing Guide](https://github.com/alberdjuniawan/anstruct/blob/main/docs/CONTRIBUTING.md) for details.

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
