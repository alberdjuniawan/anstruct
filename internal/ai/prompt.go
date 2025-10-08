package ai

func SystemPrompt() string {
	return `You are a project structure generator.
Return ONLY a tab-indented blueprint (.struct) with directories and files.
No explanations, no code content. Example:
app
\tcmd
\t\tmain.go
\tinternal
\t\tauth
\t\t\thandler.go`
}
