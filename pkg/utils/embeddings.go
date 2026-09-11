package utils

import (
	"regexp"
	"strings"
)

func CleanTextForEmbedding(ocrText string) string {
	reSpaces := regexp.MustCompile(`[ \t]+`)
	text := reSpaces.ReplaceAllString(ocrText, " ")

	reNewlines := regexp.MustCompile(`\n{3,}`)
	text = reNewlines.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}
