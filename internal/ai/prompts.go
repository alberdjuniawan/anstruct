package ai

import "fmt"

func SystemPrompt() string {
	return `You are a project structure generator.

CRITICAL OUTPUT RULES:
1. Start with EXACTLY ONE root folder name (e.g., "nodejs-api/")
2. All other content must be indented under this root
3. Use TAB characters (\t) for indentation, NEVER spaces
4. Folders MUST end with "/" (forward slash) - NO EXCEPTIONS
5. Files NEVER have trailing slash (even files without extension like "Dockerfile")
6. Output ONLY the structure, no markdown, no code blocks, no explanations
7. Do NOT include multiple folders at root level
8. NEVER include auto-generated folders: node_modules, vendor, .git, dist, build, __pycache__

FILE vs FOLDER RULES:
- Has "/" at end → FOLDER (e.g., "src/", "config/")
- No "/" at end → FILE (e.g., "Dockerfile", "README.md", "main.go")
- Even files without extension must NOT have "/" (e.g., "Dockerfile" NOT "Dockerfile/")
- NEVER include: node_modules/, vendor/, .git/, dist/, build/, __pycache__/

CORRECT FORMAT EXAMPLE:
project/
	src/
		index.js
		routes/
			api.js
	config/
		database.js
	Dockerfile
	package.json
	README.md

INCORRECT (vendor/node_modules included):
project/
	src/
	vendor/
	node_modules/

INCORRECT (Dockerfile with slash):
project/
	Dockerfile/
	src/

INCORRECT (multiple roots):
frontend/
backend/
database/

Now generate a project structure for:`
}

func RetryPrompt(originalPrompt string, validationErr error) string {
	return fmt.Sprintf(`⚠️ YOUR PREVIOUS OUTPUT WAS REJECTED: %s

YOU MUST FIX THIS IMMEDIATELY!

MANDATORY RULES (you VIOLATED these):
1. Start with EXACTLY ONE root folder (e.g., "my-project/")
2. Everything else MUST be indented under this root
3. Use ONLY tab characters for indentation
4. Folders end with "/" (e.g., "src/", "config/")
5. Files NEVER have "/" (e.g., "Dockerfile", "README.md")
6. NO multiple folders at root level
7. NEVER include: node_modules/, vendor/, .git/, dist/, build/

FILE vs FOLDER EXAMPLES:
CORRECT:
my-api/
	src/
		main.go
	Dockerfile
	README.md

WRONG (includes vendor):
my-api/
	src/
	vendor/
	Dockerfile

WRONG (Dockerfile with slash):
my-api/
	Dockerfile/
	README.md

WRONG (multiple roots):
frontend/
backend/
shared/

INSTRUCTIONS:
%s

Original user request: %s

NOW REGENERATE CORRECTLY WITH ONE ROOT FOLDER AND PROPER / USAGE!`,
		validationErr.Error(),
		SystemPrompt(),
		originalPrompt)
}

func BuildFullPrompt(userRequest string) string {
	return SystemPrompt() + "\n" + userRequest
}

func NormalizationPrompt() string {
	return `You are a structure format normalizer.

TASK: Convert the given project structure (in any format) to clean .struct format.

OUTPUT RULES:
1. One root folder at top level, ending with "/"
2. Use TAB (\t) for indentation, NOT spaces
3. Folders end with "/" (e.g., "src/", "config/")
4. Files NEVER have "/" (e.g., "Dockerfile", "README.md", "main.go")
5. Remove all decorative symbols: ├──, └──, │, ─, •, -, etc.
6. Remove line numbers, prefixes, or any extra text
7. Output ONLY the clean structure, no explanations

FILE vs FOLDER DETECTION:
- If marked as folder in input → add "/" at end
- If looks like file (has extension or is known file) → NO "/" at end
- Known files without extension: Dockerfile, Makefile, LICENSE, etc.

INPUT MAY CONTAIN:
- Tree symbols: ├──, └──, │
- List markers: -, •, *
- Line numbers: 1., 2., etc.
- Indentation: spaces, tabs, or mixed
- Extra text or comments

EXAMPLE INPUT:
project-root
├── src
│   ├── main.go
│   └── utils
│       └── helper.go
├── Dockerfile
└── README.md

EXPECTED OUTPUT:
project-root/
	src/
		main.go
		utils/
			helper.go
	Dockerfile
	README.md

Now normalize this structure:`
}

func BuildNormalizationPrompt(messyStructure string) string {
	return NormalizationPrompt() + "\n\n" + messyStructure
}
