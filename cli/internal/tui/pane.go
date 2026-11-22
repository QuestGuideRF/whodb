package tui
import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)
type Pane interface {
	UpdatePane(msg tea.Msg) tea.Cmd
	View() string
	SetDimensions(width, height int)
	Focusable() bool
	OnFocus()
	OnBlur()
	SetCompact(compact bool)
	HelpBindings() []key.Binding
}
type PaneID int
const (
	PaneLeft PaneID = iota
	PaneTopRight
	PaneBottomRight
	PaneCenter
	PaneFull
)