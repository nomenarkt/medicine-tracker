package util

import "strings"

// EscapeMarkdown prepares strings for safe MarkdownV2 output.
func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"`", "\\`",
		".", "\\.",
		"-", "\\-",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
