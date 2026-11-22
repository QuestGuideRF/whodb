package tui
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clidey/whodb/cli/pkg/styles"
)
type JSONViewer struct {
	parent     *MainModel
	viewport   viewport.Model
	columnName string
	rawContent string
	width      int
	height     int
	ready      bool
}
func NewJSONViewer(parent *MainModel) *JSONViewer {
	return &JSONViewer{
		parent: parent,
	}
}
func (v *JSONViewer) SetContent(columnName, content string) {
	v.columnName = columnName
	v.rawContent = content
	v.ready = false
}
func (v *JSONViewer) Update(msg tea.Msg) (*JSONViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.initViewport()
		return v, nil
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Global.Back) {
			if !v.parent.PopView() {
				v.parent.mode = ViewResults
			}
			return v, nil
		}
	}
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}
func (v *JSONViewer) View() string {
	if !v.ready {
		v.initViewport()
	}
	var b strings.Builder
	title := "Cell Viewer"
	if v.columnName != "" {
		title += " — " + v.columnName
	}
	b.WriteString(styles.RenderTitle(title))
	b.WriteString("\n\n")
	b.WriteString(v.viewport.View())
	b.WriteString("\n\n")
	b.WriteString(renderFooterHelpPairsWidth(v.width,
		"↑/↓", "scroll",
		"esc", "close",
	))
	if v.viewport.TotalLineCount() > v.viewport.VisibleLineCount() {
		pct := v.viewport.ScrollPercent() * 100
		var scrollPct string
		if pct >= 99.5 {
			scrollPct = "bottom"
		} else {
			scrollPct = formatFloat(pct) + "%"
		}
		b.WriteString("  ")
		b.WriteString(styles.RenderMuted(scrollPct))
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
func (v *JSONViewer) initViewport() {
	contentWidth := v.width - 8
	if contentWidth < 20 {
		contentWidth = 20
	}
	contentHeight := v.height - 10
	if contentHeight < 3 {
		contentHeight = 3
	}
	v.viewport = viewport.New(contentWidth, contentHeight)
	v.viewport.SetContent(highlightJSON(v.rawContent))
	v.ready = true
}
func highlightJSON(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return styles.MutedStyle.Render("(empty)")
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(trimmed), "", "  "); err != nil {
		return raw
	}
	return colorizeJSON(buf.String())
}
func colorizeJSON(formatted string) string {
	var out strings.Builder
	runes := []rune(formatted)
	i := 0
	for i < len(runes) {
		ch := runes[i]
		switch {
		case ch == '"':
			str, end := consumeJSONString(runes, i)
			isKey := false
			for j := end; j < len(runes); j++ {
				if runes[j] == ':' {
					isKey = true
					break
				}
				if !unicode.IsSpace(runes[j]) {
					break
				}
			}
			if isKey {
				out.WriteString(styles.KeywordStyle.Render(str))
			} else {
				out.WriteString(styles.StringStyle.Render(str))
			}
			i = end
		case ch == '-' || (ch >= '0' && ch <= '9'):
			start := i
			for i < len(runes) && isNumberRune(runes[i]) {
				i++
			}
			out.WriteString(styles.NumberStyle.Render(string(runes[start:i])))
		case ch == 't' || ch == 'f' || ch == 'n':
			word, end := consumeKeyword(runes, i)
			if word == "true" || word == "false" || word == "null" {
				out.WriteString(lipgloss.NewStyle().Foreground(styles.Info).Render(word))
				i = end
			} else {
				out.WriteRune(ch)
				i++
			}
		default:
			out.WriteRune(ch)
			i++
		}
	}
	return out.String()
}
func consumeJSONString(runes []rune, i int) (string, int) {
	j := i + 1
	for j < len(runes) {
		if runes[j] == '\\' {
			j += 2
			continue
		}
		if runes[j] == '"' {
			j++
			return string(runes[i:j]), j
		}
		j++
	}
	return string(runes[i:]), len(runes)
}
func consumeKeyword(runes []rune, i int) (string, int) {
	j := i
	for j < len(runes) && ((runes[j] >= 'a' && runes[j] <= 'z') || (runes[j] >= 'A' && runes[j] <= 'Z')) {
		j++
	}
	return string(runes[i:j]), j
}
func isNumberRune(r rune) bool {
	return (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '+' || r == 'e' || r == 'E'
}
func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", f), "0"), ".")
}