## Contributing

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
