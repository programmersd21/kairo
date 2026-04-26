package ai_panel

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/programmersd21/kairo/internal/ai"
	"github.com/programmersd21/kairo/internal/ui/styles"
	"google.golang.org/genai"
)

type AIChunkMsg struct {
	Chunk ai.StreamChunk
}

type Model struct {
	Styles   Styles
	Width    int
	Height   int
	Visible  bool
	History  []*genai.Content
	Viewport viewport.Model
	Input    textinput.Model
	Spinner  spinner.Model
	Loading  bool
	buf      *strings.Builder // accumulates raw chat text (unstyled)
	wrapW    int              // viewport text width for word-wrapping
}

func New(s styles.Styles) Model {
	ti := textinput.New()
	ti.Placeholder = "Ask Gemini..."
	ti.Focus()
	ti.CharLimit = 2000

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(s.Theme.Accent)

	return Model{
		Styles:  DefaultStyles(s),
		Input:   ti,
		Spinner: sp,
		History: []*genai.Content{},
		wrapW:   40,
		buf:     &strings.Builder{},
	}
}

func (m *Model) SetStyles(s styles.Styles) {
	m.Styles = DefaultStyles(s)
	m.Spinner.Style = lipgloss.NewStyle().Foreground(s.Theme.Accent)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.Input.Focus(), m.Spinner.Tick)
}

func (m *Model) Toggle() {
	m.Visible = !m.Visible
}

func (m *Model) SetSize(w, h int) {
	m.Width = int(float64(w) * 0.4)
	m.Height = h
	m.wrapW = m.Width - 4
	if m.wrapW < 1 {
		m.wrapW = 1
	}
	m.Input.Width = m.wrapW - 6
	m.Viewport = viewport.New(m.wrapW, m.Height-7)
}

// SetSizeExact sets the panel to exact dimensions (used when the main model
// has already computed the width split). Skips viewport recreation if
// dimensions haven't changed to avoid per-frame allocation overhead.
func (m *Model) SetSizeExact(w, h int) {
	if m.Width == w && m.Height == h {
		return // nothing changed
	}
	m.Width = w
	m.Height = h
	m.wrapW = w - 4 // border + padding
	if m.wrapW < 1 {
		m.wrapW = 1
	}
	vpH := h - 7
	if vpH < 1 {
		vpH = 1
	}
	m.Input.Width = m.wrapW - 6
	if m.Input.Width < 1 {
		m.Input.Width = 1
	}
	m.Viewport = viewport.New(m.wrapW, vpH)
	m.rebuildContent()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.Visible {
			return m, nil
		}
		switch msg.String() {
		case "enter":
			if m.Input.Value() != "" && !m.Loading {
				prompt := m.Input.Value()
				m.Input.SetValue("")
				m.Loading = true
				m.appendRaw("USER", prompt)
				return m, func() tea.Msg { return prompt } // Bridge to main model
			}
		case "esc":
			m.Visible = false
			return m, nil
		}

	case AIChunkMsg:
		if msg.Chunk.Err != nil {
			m.appendRaw("ERR", msg.Chunk.Err.Error())
			m.Loading = false
		} else if msg.Chunk.Done {
			m.buf.WriteString("\n")
			m.rebuildContent()
			m.Loading = false
		} else if msg.Chunk.ToolUse != nil {
			m.appendRaw("TOOL", msg.Chunk.ToolUse.ToolName)
		} else {
			// Stream AI text — append raw, rebuild with wrapping
			m.buf.WriteString(msg.Chunk.Text)
			m.rebuildContent()
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// appendRaw adds a tagged line to the raw buffer (kind: USER, ERR, TOOL).
func (m *Model) appendRaw(kind, text string) {
	switch kind {
	case "USER":
		m.buf.WriteString("▸ " + text + "\n")
	case "TOOL":
		m.buf.WriteString("⚙ " + text + "\n")
	case "ERR":
		m.buf.WriteString("✗ " + text + "\n")
	}
	m.rebuildContent()
}

// rebuildContent re-wraps and re-styles the raw buffer text for the viewport.
func (m *Model) rebuildContent() {
	raw := m.buf.String()
	lines := strings.Split(raw, "\n")
	var out strings.Builder

	for _, line := range lines {
		if line == "" {
			out.WriteString("\n")
			continue
		}
		var styled string
		switch {
		case strings.HasPrefix(line, "▸ "):
			wrapped := lipgloss.NewStyle().Width(m.wrapW).Render(line)
			styled = m.Styles.User.Render(wrapped)
		case strings.HasPrefix(line, "⚙ "):
			wrapped := lipgloss.NewStyle().Width(m.wrapW).Render(line)
			styled = m.Styles.Tool.Render(wrapped)
		case strings.HasPrefix(line, "✗ "):
			wrapped := lipgloss.NewStyle().Width(m.wrapW).Render(line)
			styled = m.Styles.Tool.Render(wrapped)
		default:
			// AI response text — wrap to viewport width
			styled = m.Styles.AI.Width(m.wrapW).Render(line)
		}
		out.WriteString(styled + "\n")
	}

	m.Viewport.SetContent(out.String())
	m.Viewport.GotoBottom()
}

func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	innerW := m.wrapW
	if innerW < 1 {
		innerW = 1
	}

	// ── Header ──
	headerText := " 🤖 Kairo AI "
	if m.Loading {
		headerText = " " + m.Spinner.View() + " Thinking… "
	}
	header := m.Styles.Header.Width(innerW).Render(headerText)

	// ── Chat viewport ──
	chat := m.Viewport.View()

	// ── Input area ──
	inputBox := m.Styles.InputBorder.Width(innerW - 4).Render(m.Input.View())

	// ── Footer ──
	footer := m.Styles.Footer.Width(innerW).Align(lipgloss.Center).
		Render("esc close · enter send")

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		chat,
		"",
		inputBox,
		footer,
	)

	return m.Styles.Panel.Width(m.Width).Height(m.Height).Render(content)
}
