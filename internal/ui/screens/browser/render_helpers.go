package browser

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

type columnLayout struct {
	indexes       []int
	widths        []int
	hiddenColumns int
	firstColumn   int
	lastColumn    int
	scrollOffset  int
	maxOffset     int
}

func fitCell(value string, width int) string {
	if width <= 1 {
		return value
	}

	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > width {
		return padCell(string(runes[:width-1])+"…", width)
	}
	return padCell(value, width)
}

func padCell(value string, width int) string {
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(value)
}

func runeLen(value string) int {
	return utf8.RuneCountInString(value)
}

func visibleLen(value string) int {
	return runeLen(strings.ReplaceAll(stripANSI(value), "\n", ""))
}

func stripANSI(value string) string {
	var out []rune
	inEscape := false
	for _, r := range value {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

func wrapText(value string, width int) []string {
	if width <= 1 {
		return []string{value}
	}

	if value == "" {
		return []string{""}
	}

	words := strings.Fields(value)
	if len(words) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, 4)
	current := words[0]

	for _, word := range words[1:] {
		candidate := current + " " + word
		if runeLen(candidate) <= width {
			current = candidate
			continue
		}

		if runeLen(current) > width {
			lines = append(lines, hardWrap(current, width)...)
		} else {
			lines = append(lines, current)
		}
		current = word
	}

	if runeLen(current) > width {
		lines = append(lines, hardWrap(current, width)...)
	} else {
		lines = append(lines, current)
	}

	return lines
}

func hardWrap(value string, width int) []string {
	if width <= 1 {
		return []string{value}
	}

	runes := []rune(value)
	lines := make([]string, 0, (len(runes)/width)+1)
	for len(runes) > width {
		lines = append(lines, string(runes[:width]))
		runes = runes[width:]
	}
	if len(runes) > 0 {
		lines = append(lines, string(runes))
	}
	return lines
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
