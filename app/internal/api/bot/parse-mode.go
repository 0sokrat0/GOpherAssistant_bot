package bot

import (
	tb "gopkg.in/telebot.v4"
	"regexp"
	"strings"
)

// Экранирование текста для MarkdownV2
func escapeMarkdownV2(text string) string {
	return strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"`", "\\`",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	).Replace(text)
}

// Регулярное выражение для блоков кода
var codeBlockRegex = regexp.MustCompile("(?s)```(.*?)```")

func formatResponse(response string) string {
	var result strings.Builder
	lastIndex := 0

	// Ищем блоки кода
	matches := codeBlockRegex.FindAllStringSubmatchIndex(response, -1)
	for _, match := range matches {
		// Добавляем текст перед кодом
		if lastIndex < match[0] {
			plainText := response[lastIndex:match[0]]
			result.WriteString(escapeMarkdownV2(plainText))
		}

		// Добавляем код в формате Markdown
		codeBlock := response[match[2]:match[3]]
		result.WriteString("```\n" + strings.TrimSpace(codeBlock) + "\n```\n")

		lastIndex = match[1]
	}

	// Добавляем оставшийся текст после последнего блока кода
	if lastIndex < len(response) {
		plainText := response[lastIndex:]
		result.WriteString(escapeMarkdownV2(plainText))
	}

	return result.String()
}

func splitMessage(msg string, limit int) []string {
	var parts []string
	for len(msg) > limit {
		splitIndex := strings.LastIndex(msg[:limit], "\n")
		if splitIndex == -1 {
			splitIndex = limit
		}
		parts = append(parts, msg[:splitIndex])
		msg = msg[splitIndex:]
	}
	parts = append(parts, msg)
	return parts
}

func sendLongMessage(c tb.Context, message string) error {
	parts := splitMessage(message, 4000) // Telegram позволяет до 4096 символов
	for _, part := range parts {
		if err := c.Send(part, tb.ModeMarkdownV2); err != nil {
			return err
		}
	}
	return nil
}
