package tui
import (
	"fmt"
	"strings"
	"time"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clidey/whodb/cli/internal/database"
	"github.com/clidey/whodb/cli/pkg/styles"
)
type CmdLogView struct {
	parent   *MainModel
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}
func NewCmdLogView(parent *MainModel) *CmdLogView {
	return &CmdLogView{
		parent: parent,
	}
}
func (v *CmdLogView) Update(msg tea.Msg) (*CmdLogView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.ready = false
		return v, nil
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Global.Back) {
			if !v.parent.PopView() {
				v.parent.mode = ViewBrowser
			}
			return v, nil
		}
	}
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}
func (v *CmdLogView) View() string {
	v.initViewport()
	var b strings.Builder
	b.WriteString(styles.RenderTitle("Command Log"))
	b.WriteString("\n\n")
	b.WriteString(v.viewport.View())
	b.WriteString("\n\n")
	helpText := RenderBindingHelpWidth(v.width,
		key.NewBinding(key.WithKeys("up/down"), key.WithHelp("up/down", "scroll")),
		Keys.Global.Back,
	)
	b.WriteString(helpText)
	if v.viewport.TotalLineCount() > v.viewport.VisibleLineCount() {
		pct := v.viewport.ScrollPercent() * 100
		var scrollPct string
		if pct >= 99.5 {
			scrollPct = "bottom"
		} else {
			scrollPct = fmt.Sprintf("%.0f%%", pct)
		}
		b.WriteString("  ")
		b.WriteString(styles.RenderMuted(scrollPct))
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
func (v *CmdLogView) initViewport() {
	contentWidth := v.width - 8
	if contentWidth < 20 {
		contentWidth = 20
	}
	contentHeight := v.height - 10
	if contentHeight < 3 {
		contentHeight = 3
	}
	if !v.ready {
		v.viewport = viewport.New(contentWidth, contentHeight)
		v.ready = true
	} else {
		v.viewport.Width = contentWidth
		v.viewport.Height = contentHeight
	}
	v.viewport.SetContent(v.renderLogEntries())
}
func (v *CmdLogView) renderLogEntries() string {
	entries := v.parent.dbManager.GetQueryLog()
	if len(entries) == 0 {
		return styles.MutedStyle.Render("No queries executed yet.")
	}
	maxQueryWidth := v.width - 50
	if maxQueryWidth < 20 {
		maxQueryWidth = 20
	}
	var b strings.Builder
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		b.WriteString(renderLogEntry(entry, maxQueryWidth))
		if i > 0 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
func renderLogEntry(entry database.QueryLogEntry, maxQueryWidth int) string {
	timestamp := styles.MutedStyle.Render(fmt.Sprintf("[%s]", entry.Timestamp.Format("15:04:05")))
	query := strings.ReplaceAll(entry.Query, "\n", " ")
	query = strings.Join(strings.Fields(query), " ")
	if len(query) > maxQueryWidth {
		query = query[:maxQueryWidth-3] + "..."
	}
	durationStr := formatDuration(entry.Duration)
	var status string
	if entry.Success {
		rowInfo := ""
		if entry.RowCount > 0 {
			rowInfo = fmt.Sprintf(", %d rows", entry.RowCount)
		}
		status = styles.SuccessStyle.Render(fmt.Sprintf("(%s%s) OK", durationStr, rowInfo))
	} else {
		errMsg := entry.Error
		if len(errMsg) > 40 {
			errMsg = errMsg[:37] + "..."
		}
		status = styles.ErrorStyle.Render(fmt.Sprintf("(%s) ERR %s", durationStr, errMsg))
	}
	return fmt.Sprintf("%s %s  %s", timestamp, query, status)
}
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dus", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}