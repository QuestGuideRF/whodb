package layout
import (
	"strings"
	"github.com/charmbracelet/lipgloss"
	"github.com/clidey/whodb/cli/pkg/styles"
)
type Renderable interface {
	View() string
	SetDimensions(width, height int)
}
type Direction int
const (
	Horizontal Direction = iota
	Vertical
)
type Container struct {
	direction Direction
	ratio     float64
	children  [2]*Container
	content   Renderable
	label     string
	focused   bool
	x, y, w, h int
}
const MinPaneWidth = 20
const MinPaneHeight = 4
const borderOverhead = 1
func NewLeaf(label string, content Renderable) *Container {
	return &Container{
		content: content,
		label:   label,
		ratio:   0.5,
	}
}
func NewSplit(dir Direction, ratio float64, first, second *Container) *Container {
	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}
	return &Container{
		direction: dir,
		ratio:     ratio,
		children:  [2]*Container{first, second},
	}
}
func (c *Container) IsLeaf() bool {
	return c.content != nil
}
func (c *Container) Label() string {
	return c.label
}
func (c *Container) SetFocused(focused bool) {
	c.focused = focused
}
func (c *Container) IsFocused() bool {
	return c.focused
}
func (c *Container) Content() Renderable {
	return c.content
}
func (c *Container) Children() [2]*Container {
	return c.children
}
func (c *Container) Geometry() (x, y, w, h int) {
	return c.x, c.y, c.w, c.h
}
func (c *Container) Ratio() float64 {
	return c.ratio
}
func (c *Container) SetRatio(ratio float64) {
	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}
	c.ratio = ratio
}
func (c *Container) AdjustRatio(delta float64) {
	c.SetRatio(c.ratio + delta)
}
func (c *Container) Direction() Direction {
	return c.direction
}
func (c *Container) Layout(x, y, w, h int) {
	c.x, c.y, c.w, c.h = x, y, w, h
	if c.IsLeaf() {
		contentH := h
		if contentH < 1 {
			contentH = 1
		}
		c.content.SetDimensions(w, contentH)
		return
	}
	first, second := c.children[0], c.children[1]
	if first == nil || second == nil {
		return
	}
	switch c.direction {
	case Horizontal:
		divider := 1
		available := w - divider
		if available < MinPaneWidth*2 {
			first.Layout(x, y, w, h)
			second.Layout(x, y, 0, 0)
			return
		}
		firstW := int(float64(available) * c.ratio)
		if firstW < MinPaneWidth {
			firstW = MinPaneWidth
		}
		secondW := available - firstW
		if secondW < MinPaneWidth {
			secondW = MinPaneWidth
			firstW = available - secondW
		}
		first.Layout(x, y, firstW, h)
		second.Layout(x+firstW+divider, y, secondW, h)
	case Vertical:
		divider := 1
		available := h - divider
		if available < MinPaneHeight*2 {
			first.Layout(x, y, w, h)
			second.Layout(x, y, 0, 0)
			return
		}
		firstH := int(float64(available) * c.ratio)
		if firstH < MinPaneHeight {
			firstH = MinPaneHeight
		}
		secondH := available - firstH
		if secondH < MinPaneHeight {
			secondH = MinPaneHeight
			firstH = available - secondH
		}
		first.Layout(x, y, w, firstH)
		second.Layout(x, y+firstH+divider, w, secondH)
	}
}
func (c *Container) View() string {
	if c.w <= 0 || c.h <= 0 {
		return ""
	}
	if c.IsLeaf() {
		return c.renderLeaf()
	}
	first, second := c.children[0], c.children[1]
	if first == nil || second == nil {
		return ""
	}
	firstView := first.View()
	secondView := second.View()
	if first.w <= 0 || first.h <= 0 {
		return secondView
	}
	if second.w <= 0 || second.h <= 0 {
		return firstView
	}
	switch c.direction {
	case Horizontal:
		divider := renderVerticalDivider(c.h)
		return lipgloss.JoinHorizontal(lipgloss.Top, firstView, divider, secondView)
	case Vertical:
		divider := renderHorizontalDivider(c.w)
		return lipgloss.JoinVertical(lipgloss.Left, firstView, divider, secondView)
	}
	return ""
}
func (c *Container) renderLeaf() string {
	headerH := borderOverhead
	contentH := c.h - headerH
	if contentH < 1 {
		contentH = 1
	}
	c.content.SetDimensions(c.w, contentH)
	raw := c.content.View()
	content := fitToBox(raw, c.w, contentH)
	header := renderPaneHeader(c.label, c.w, c.focused)
	return header + "\n" + content
}
func (c *Container) ViewSingle() string {
	if !c.IsLeaf() {
		return c.View()
	}
	c.content.SetDimensions(c.w, c.h)
	return c.content.View()
}
func (c *Container) Leaves() []*Container {
	if c.IsLeaf() {
		return []*Container{c}
	}
	var leaves []*Container
	if c.children[0] != nil {
		leaves = append(leaves, c.children[0].Leaves()...)
	}
	if c.children[1] != nil {
		leaves = append(leaves, c.children[1].Leaves()...)
	}
	return leaves
}
func renderPaneHeader(label string, width int, focused bool) string {
	if width <= 0 {
		return ""
	}
	color := styles.Muted
	if focused {
		color = styles.Primary
	}
	lineChar := "─"
	style := lipgloss.NewStyle().Foreground(color)
	if label == "" {
		return style.Render(strings.Repeat(lineChar, width))
	}
	labelStr := " " + label + " "
	labelWidth := lipgloss.Width(labelStr)
	remaining := width - labelWidth
	if remaining < 2 {
		return style.Render(strings.Repeat(lineChar, width))
	}
	leftLine := strings.Repeat(lineChar, 1)
	rightLine := strings.Repeat(lineChar, remaining-1)
	return style.Render(leftLine + labelStr + rightLine)
}
func renderVerticalDivider(height int) string {
	if height <= 0 {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(styles.Border)
	lines := make([]string, height)
	for i := range lines {
		lines[i] = style.Render("│")
	}
	return strings.Join(lines, "\n")
}
func renderHorizontalDivider(width int) string {
	if width <= 0 {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(styles.Border)
	return style.Render(strings.Repeat("─", width))
}
func fitToBox(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		w := lipgloss.Width(line)
		if w > width {
			lines[i] = truncateLine(line, width)
		} else if w < width {
			lines[i] = line + strings.Repeat(" ", width-w)
		}
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return strings.Join(lines, "\n")
}
func truncateLine(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(s)
}