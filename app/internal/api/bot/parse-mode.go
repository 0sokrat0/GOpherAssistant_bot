package bot

import (
	"fmt"
	"regexp"
	"strings"
)

// Регулярные выражения для распознавания форматов
var (
	codeBlockRegex = regexp.MustCompile("(?s)```(.*?)```")          // Кодовые блоки
	listRegex      = regexp.MustCompile(`(?m)^(?:\*|-|\d+\.)\s+.*`) // Списки
	headingRegex   = regexp.MustCompile(`(?m)^#{1,6}\s+.*`)         // Заголовки
	quoteRegex     = regexp.MustCompile(`(?m)^>.*`)                 // Цитаты
	linkRegex      = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)  // Ссылки
)

func formatGPTResponse(response string, useMarkdownV2 bool) string {
	var result strings.Builder
	lastIndex := 0

	// Форматирование блоков кода
	matches := codeBlockRegex.FindAllStringSubmatchIndex(response, -1)
	for _, match := range matches {
		// Текст перед блоком кода
		if lastIndex < match[0] {
			plainText := response[lastIndex:match[0]]
			result.WriteString(formatText(plainText, useMarkdownV2))
		}

		// Форматируем блок кода
		codeBlock := response[match[2]:match[3]]
		result.WriteString(formatCodeBlock(codeBlock, useMarkdownV2))

		lastIndex = match[1]
	}

	// Оставшийся текст после последнего блока кода
	if lastIndex < len(response) {
		plainText := response[lastIndex:]
		result.WriteString(formatText(plainText, useMarkdownV2))
	}

	return result.String()
}

func formatText(text string, useMarkdownV2 bool) string {
	if useMarkdownV2 {
		return escapeMarkdownV2(strings.TrimSpace(text)) + "\n"
	}
	return strings.TrimSpace(text) + "\n"
}

func formatCodeBlock(code string, useMarkdownV2 bool) string {
	if useMarkdownV2 {
		return "```\n" + escapeMarkdownV2(strings.TrimSpace(code)) + "\n```\n"
	}
	return "```\n" + strings.TrimSpace(code) + "\n```\n"
}

func formatList(text string, useMarkdownV2 bool) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder

	for _, line := range lines {
		if useMarkdownV2 {
			result.WriteString("- " + escapeMarkdownV2(strings.TrimSpace(line)) + "\n")
		} else {
			result.WriteString("- " + strings.TrimSpace(line) + "\n")
		}
	}

	return result.String()
}

func formatHeading(text string, useMarkdownV2 bool) string {
	if useMarkdownV2 {
		return "*_" + escapeMarkdownV2(strings.TrimSpace(text)) + "_*\n"
	}
	return strings.TrimSpace(text) + "\n"
}

func formatLink(match []string, useMarkdownV2 bool) string {
	if useMarkdownV2 {
		return fmt.Sprintf("[%s](%s)", escapeMarkdownV2(match[1]), escapeMarkdownV2(match[2]))
	}
	return fmt.Sprintf("[%s](%s)", match[1], match[2])
}

func escapeMarkdownV2(text string) string {
	return strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
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
