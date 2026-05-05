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
//   - FillViewport: Pads every line to full width and ensures background
//     color fills the entire terminal using erase-to-end-of-line (\x1b[K).
//
// Together, they guarantee that every cell in the viewport has the
// intended background color, regardless of what individual components render.
package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// bgAnsi returns the raw ANSI escape sequence for setting a 24-bit
// background color from a hex color string (e.g., "#1E1E2E").
func bgAnsi(c lipgloss.Color) string {
	if c == "" {
		return ""
	}
	// Use lipgloss to generate the ANSI sequence for this color.
	// We render an empty string with the background color set.
	// This returns something like "\x1b[48;2;...m\x1b[0m".
	s := lipgloss.NewStyle().Background(c).Render("")
	// Strip the reset sequence (\x1b[0m) if it exists.
	return strings.TrimSuffix(s, "\x1b[0m")
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
// It performs four operations:
//  1. Strips stray \r characters for cross-platform robustness
//  2. Right-pads every existing line to full width with background-colored spaces
//  3. Appends background-colored empty lines to reach the target height
//  4. Applies PaintBg to the entire result to fix internal ANSI resets
//
// After PaintBg, each line is also suffixed with the "erase to end of line"
// sequence (\x1b[K) which tells the terminal to fill the remainder of the
// line with the currently active background color. This catches any cells
// that lipgloss.Width might have miscounted.
//
// This is the definitive viewport-filling function. Apply it as the LAST step
// in the top-level View() to guarantee a seamless, gap-free background.
func FillViewport(content string, width, height int, bg lipgloss.Color) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Strip \r to normalize line endings across platforms.
	content = strings.ReplaceAll(content, "\r", "")

	lines := strings.Split(content, "\n")
	bgStyle := lipgloss.NewStyle().Background(bg)
	// Pad an extra column to guard against off-by-one width calculations
	// in some terminals/lipgloss width computations. This extra column
	// ensures the right-most cell is always painted.
	emptyLine := bgStyle.Render(strings.Repeat(" ", width+1))

	result := make([]string, 0, height)
	for i := 0; i < height; i++ {
		if i < len(lines) {
			line := lines[i]
			vis := lipgloss.Width(line)
			if vis < width {
				pad := bgStyle.Render(strings.Repeat(" ", width-vis))
				result = append(result, line+pad)
			} else if vis > width {
				// Truncate line if it's too wide to prevent terminal wrapping
				// which would break the vertical layout and cause bleeding.
				result = append(result, lipgloss.NewStyle().MaxWidth(width).Render(line))
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, emptyLine)
		}
	}

	// Append one extra full-width background line to ensure the bottom
	// row is painted in terminals that may otherwise leave a thin strip.
	// This is a harmless extra line when using the alternate screen buffer.
	result = append(result, emptyLine)

	seq := bgAnsi(bg)
	painted := PaintBg(strings.Join(result, "\n"), bg)

	// Append the "erase in line" (EL) sequence after each line.
	// \x1b[K clears from the cursor to the end of the line using the
	// currently active background color. This is the definitive fix for
	// any cells beyond the painted width—even if lipgloss.Width miscounted.
	if seq != "" {
		painted = strings.ReplaceAll(painted, "\n", "\x1b[K\n")
		// Also clear to end-of-line on the very last line, and then clear
		// the rest of the display. This ensures any terminal padding or
		// margins are filled with our background color.
		painted += "\x1b[K\x1b[J"
	}

	return painted
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
