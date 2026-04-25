// Package render provides terminal rendering utilities that guarantee
// background color consistency across the entire viewport.
//
// The core problem: Lip Gloss styled text emits ANSI SGR reset sequences
// (\x1b[0m) at the end of each styled segment. These resets clear ALL
// formatting—including background color—causing the terminal's default
// background to bleed through between styled segments.
//
// This package solves it with two complementary strategies:
//   - PaintBg: Re-applies the background color after every ANSI reset
//   - FillViewport: Pads every line to full width and fills missing lines
//
// Together, they guarantee that every cell in the viewport has the
// intended background color, regardless of what individual components render.
package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// bgAnsi returns the raw ANSI escape sequence for setting a 24-bit
// background color from a hex color string (e.g., "#1E1E2E").
func bgAnsi(c lipgloss.Color) string {
	hex := strings.TrimPrefix(string(c), "#")
	if len(hex) != 6 {
		return ""
	}
	var r, g, b int
	_, _ = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

// PaintBg is a post-processing step that ensures a background color persists
// through all ANSI SGR reset sequences in a rendered string.
//
// When Lip Gloss renders styled text, each segment ends with \x1b[0m which
// clears ALL formatting, including background. Any characters after the reset
// (spacers, padding from JoinHorizontal, etc.) inherit the terminal's default
// background instead of the app's theme background.
//
// PaintBg fixes this by:
//  1. Prepending the background ANSI sequence (so the string starts in the correct state)
//  2. Inserting the background sequence after every \x1b[0m reset
//
// This ensures no character in the string can ever fall back to the terminal default.
func PaintBg(s string, bg lipgloss.Color) string {
	seq := bgAnsi(bg)
	if seq == "" {
		return s
	}
	// Re-apply background after every SGR reset
	s = strings.ReplaceAll(s, "\x1b[0m", "\x1b[0m"+seq)
	// Ensure background is set from the very start
	return seq + s
}

// FillViewport ensures a rendered string covers exactly width×height terminal
// cells, with every unfilled cell receiving the specified background color.
//
// It performs three operations:
//  1. Right-pads every existing line to full width with background-colored spaces
//  2. Appends background-colored empty lines to reach the target height
//  3. Applies PaintBg to the entire result to fix internal ANSI resets
//
// This is the definitive viewport-filling function. Apply it as the LAST step
// in the top-level View() to guarantee a seamless, gap-free background.
func FillViewport(content string, width, height int, bg lipgloss.Color) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	lines := strings.Split(content, "\n")
	bgStyle := lipgloss.NewStyle().Background(bg)
	emptyLine := bgStyle.Render(strings.Repeat(" ", width))

	result := make([]string, 0, height)
	for i := 0; i < height; i++ {
		if i < len(lines) {
			line := lines[i]
			vis := lipgloss.Width(line)
			if vis < width {
				pad := bgStyle.Render(strings.Repeat(" ", width-vis))
				result = append(result, line+pad)
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, emptyLine)
		}
	}

	return PaintBg(strings.Join(result, "\n"), bg)
}

// BarLine creates a single full-width line with left-aligned and right-aligned
// content, filling the gap between them with background-colored spaces.
// Both left and right content retain their existing ANSI styling.
func BarLine(left, right string, width int, bg lipgloss.Color) string {
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	gap := width - lw - rw
	if gap < 0 {
		gap = 0
	}
	spacer := lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", gap))
	line := left + spacer + right
	return lipgloss.NewStyle().MaxWidth(width).MaxHeight(1).Render(line)
}
