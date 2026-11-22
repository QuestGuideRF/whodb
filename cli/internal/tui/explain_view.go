package tui
import (
	"regexp"
	"strconv"
	"strings"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clidey/whodb/cli/pkg/styles"
)
var costPattern = regexp.MustCompile(`cost=[\d.]+\.\.([\d.]+)`)
type ExplainView struct {
	parent   *MainModel
	viewport viewport.Model
	query    string
	plan     string
	width    int
	height   int
	ready    bool
}
func NewExplainView(parent *MainModel) *ExplainView {
	return &ExplainView{
		parent: parent,
	}
}
func (v *ExplainView) Update(msg tea.Msg) (*ExplainView, tea.Cmd) {
	switch msg := msg.(type) {
	case explainResultMsg:
		if msg.err != nil {
			v.SetPlan(msg.query, "Error: "+msg.err.Error())
		} else {
			v.SetPlan(msg.query, msg.plan)
		}
		return v, nil
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.initViewport()
		return v, nil
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Global.Back) {
			if !v.parent.PopView() {
				v.parent.mode = ViewEditor
			}
			return v, nil
		}
	}
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}
func (v *ExplainView) View() string {
	if !v.ready {
		v.initViewport()
	}
	var b strings.Builder
	b.WriteString(styles.RenderTitle("Query Plan"))
	b.WriteString("\n\n")
	if v.query != "" {
		truncated := v.query
		maxLen := v.width - 12
		if maxLen < 20 {
			maxLen = 20
		}
		if len(truncated) > maxLen {
			truncated = truncated[:maxLen] + "..."
		}
		b.WriteString(styles.RenderMuted(truncated))
		b.WriteString("\n\n")
	}
	b.WriteString(v.viewport.View())
	b.WriteString("\n\n")
	b.WriteString(RenderBindingHelpWidth(v.width,
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("up/dn", "scroll")),
		Keys.Global.Back,
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
func (v *ExplainView) SetPlan(query, plan string) {
	v.query = query
	v.plan = plan
	v.ready = false
}
func (v *ExplainView) initViewport() {
	contentWidth := v.width - 8
	if contentWidth < 20 {
		contentWidth = 20
	}
	contentHeight := v.height - 14
	if contentHeight < 3 {
		contentHeight = 3
	}
	v.viewport = viewport.New(contentWidth, contentHeight)
	v.viewport.SetContent(colorizePlan(v.plan))
	v.ready = true
}
func colorizePlan(plan string) string {
	if strings.TrimSpace(plan) == "" {
		return styles.MutedStyle.Render("(no plan output)")
	}
	lines := strings.Split(plan, "\n")
	var out strings.Builder
	for i, line := range lines {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString(colorizePlanLine(line))
	}
	return out.String()
}
func colorizePlanLine(line string) string {
	matches := costPattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return line
	}
	cost, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return line
	}
	var color lipgloss.AdaptiveColor
	switch {
	case cost < 100:
		color = styles.Success
	case cost <= 1000:
		color = styles.Warning
	default:
		color = styles.Error
	}
	return lipgloss.NewStyle().Foreground(color).Render(line)
}