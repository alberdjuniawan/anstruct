# Anstruct - AI-Powered Project Structure Manager

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://imgshields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Anstruct is a powerful CLI tool that revolutionizes project structure management using AI. Generate, reverse-engineer, normalize, and sync project structures with simple commands.

> 🚀 Say goodbye to manual directory creation. Define your project structure in plain English or a simple blueprint, and let Anstruct build it for you.

---

## ✨ Features at a Glance

| Feature | Description | Command |
| :--- | :--- | :--- |
| AI-Powered Generation | Create complex project structures from a single natural language prompt. | aistruct |
| Blueprint System | Define, manage, and share structures using the simple .struct format. | mstruct |
| Reverse Engineering | Convert any existing project into a reusable .struct blueprint. | rstruct |
| Format Normalization | Convert tree, ls -R, or even messy markdown structures into .struct. | convert |
| Real-time Sync | Bidirectional sync between a project folder and its blueprint file. | watch |
| History & Rollback | Undo/redo project operations with full history tracking. | history |
| Performance | Written in Go for maximum speed and efficiency. | |

---

## ⬇️ Installation

### Using Go Install (Recommended)

This is the fastest way to get the latest stable release.

go install github.com/alberdjuniawan/anstruct/cmd/anstruct@latest

### From Source

For development or specific version control.

# Clone repository
git clone https://github.com/alberdjuniawan/anstruct.git
cd anstruct

# Build and install globally
go install ./cmd/anstruct

---

## ⚡ Quick Start

Experience the power of Anstruct with a few simple commands.

### 1. AI-Powered Generation

Generate an entire project structure from a prompt. Use the --apply flag to create files directly, or omit it to generate a blueprint file first.

# Generate a Go REST API blueprint
anstruct aistruct "golang REST API with PostgreSQL and Docker" -o api.struct

# Generate a React dashboard project directly
anstruct aistruct "react dashboard with routing and tailwind" --apply -o ./dashboard

Tip: Be specific in your prompt for the best results!

### 2. Reverse Engineer & Document

Convert an existing project into a blueprint for documentation, sharing, or modification.

anstruct rstruct ./my-existing-app -o app.struct

### 3. Real-time Watch & Sync

Keep your folder and blueprint in sync during development. This is perfect for ensuring team consistency.

# Two-way sync: folder <-> blueprint
anstruct watch ./myapp ./myapp.struct --full

# One-way sync: blueprint -> folder (safer for production)
anstruct watch ./myapp ./myapp.struct --half struct

---

## 📖 Comprehensive Documentation

For detailed commands, .struct specification, and advanced usage, please check the dedicated documentation files:

| Document | Content |
| :--- | :--- |
| [Commands Reference](docs/COMMANDS.md) | Full breakdown of all commands (aistruct, mstruct, rstruct, watch, etc.) and their flags. |
| [Blueprint Format (.struct)](docs/FORMAT.md) | Detailed specification of the simple, human-readable .struct file format. |
| [Advanced Usage & Best Practices](docs/ADVANCED.md) | Use cases, CI/CD integration, custom AI prompts, and optimization tips. |

---

## 🤝 Contributing

We welcome contributions! Please see our [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on submitting issues, feature requests, and pull requests.

---

## ⭐️ Acknowledgments

- Powered by [Gemini AI](https://deepmind.google/technologies/gemini/)
- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- File watching by [fsnotify](https://github.com/fsnotify/fsnotify)

---

Made by [Alberd Juniawan](https://github.com/alberdjuniawan) | [MIT License](LICENSE)

[⭐ Star on GitHub](https://github.com/alberdjuniawan/anstruct) • [Report Bug](https://github.com/alberdjuniawan/anstruct/issues) • [Request Feature](https://github.com/alberdjuniawan/anstruct/issues)
