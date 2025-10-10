package ai

import "fmt"

// SystemPrompt returns the base instruction for AI
func SystemPrompt() string {
	return `You are a project structure generator.

CRITICAL OUTPUT RULES:
1. Start with EXACTLY ONE root folder name (e.g., "nodejs-api/")
2. All other content must be indented under this root
3. Use TAB characters (\t) for indentation, NEVER spaces
4. Folders MUST end with "/" (forward slash)
5. Files have extensions, no trailing slash
6. Output ONLY the structure, no markdown, no code blocks, no explanations
7. Do NOT include multiple folders at root level

CORRECT FORMAT EXAMPLE:
project/
	src/
		index.js
		routes/
			api.js
	config/
		database.js
	package.json
	README.md

INCORRECT (multiple roots):
frontend/
backend/
database/

Now generate a project structure for:`
}

// RetryPrompt builds a correction prompt for retry attempts
func RetryPrompt(originalPrompt string, validationErr error) string {
	return fmt.Sprintf(`⚠️ YOUR PREVIOUS OUTPUT WAS REJECTED: %s

YOU MUST FIX THIS IMMEDIATELY!

MANDATORY RULES (you VIOLATED these):
1. ✅ Start with EXACTLY ONE root folder (e.g., "my-project/")
2. ✅ Everything else MUST be indented under this root
3. ✅ Use ONLY tab characters for indentation
4. ✅ Folders end with "/" 
5. ✅ NO multiple folders at root level

CORRECT EXAMPLE (COPY THIS STRUCTURE):
my-api/
	src/
		index.js
		routes/
			users.js
	package.json
	README.md

WRONG EXAMPLE (DO NOT DO THIS):
frontend/
backend/
shared/

INSTRUCTIONS:
%s

Original user request: %s

NOW REGENERATE CORRECTLY WITH ONE ROOT FOLDER ONLY!`,
		validationErr.Error(),
		SystemPrompt(),
		originalPrompt)
}

// BuildFullPrompt combines system prompt with user request
func BuildFullPrompt(userRequest string) string {
	return SystemPrompt() + "\n" + userRequest
}
