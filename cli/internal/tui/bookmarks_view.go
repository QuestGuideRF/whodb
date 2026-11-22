package tui
import (
	"fmt"
	"strings"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clidey/whodb/cli/pkg/styles"
)
const saveCurrentOption = -1
type BookmarksView struct {
	parent      *MainModel
	width       int
	height      int
	cursor      int
	editorQuery string
	naming      bool
	nameInput   textinput.Model
}
func NewBookmarksView(parent *MainModel) *BookmarksView {
	ni := textinput.New()
	ni.Placeholder = "bookmark name"
	ni.CharLimit = 50
	ni.Width = 30
	ni.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ni.TextStyle = lipgloss.NewStyle().Foreground(styles.Foreground)
	ni.Cursor.Style = lipgloss.NewStyle().Foreground(styles.Primary)
	return &BookmarksView{
		parent:    parent,
		cursor:    0,
		nameInput: ni,
	}
}
func (v *BookmarksView) itemCount() int {
	n := len(v.parent.config.GetSavedQueries())
	if v.editorQuery != "" {
		n++
	}
	return n
}
func (v *BookmarksView) isSaveRow() bool {
	return v.editorQuery != "" && v.cursor == 0
}
func (v *BookmarksView) queryIndex() int {
	if v.editorQuery != "" {
		return v.cursor - 1
	}
	return v.cursor
}
func (v *BookmarksView) nextBookmarkName() string {
	existing := v.parent.config.GetSavedQueries()
	max := 0
	for _, sq := range existing {
		var n int
		if _, err := fmt.Sscanf(sq.Name, "Bookmark %d", &n); err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("Bookmark %d", max+1)
}
func (v *BookmarksView) Update(msg tea.Msg) (*BookmarksView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			if v.cursor > 0 {
				v.cursor--
			}
			return v, nil
		case tea.MouseButtonWheelDown:
			if v.cursor < v.itemCount()-1 {
				v.cursor++
			}
			return v, nil
		}
	case tea.KeyMsg:
		if v.naming {
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(v.nameInput.Value())
				if name == "" {
					name = v.nextBookmarkName()
				}
				v.parent.config.AddSavedQuery(name, v.editorQuery)
				v.naming = false
				v.nameInput.Blur()
				v.nameInput.SetValue("")
				return v, tea.Batch(v.parent.requestConfigSave(), v.parent.SetStatus("Saved: "+name))
			case "esc":
				v.naming = false
				v.nameInput.Blur()
				v.nameInput.SetValue("")
				return v, nil
			default:
				v.nameInput, _ = v.nameInput.Update(msg)
				return v, nil
			}
		}
		switch {
		case key.Matches(msg, Keys.Global.Back):
			if !v.parent.PopView() {
				v.parent.mode = ViewEditor
			}
			return v, nil
		case key.Matches(msg, Keys.Bookmarks.Up):
			if v.cursor > 0 {
				v.cursor--
			}
			return v, nil
		case key.Matches(msg, Keys.Bookmarks.Down):
			if v.cursor < v.itemCount()-1 {
				v.cursor++
			}
			return v, nil
		case key.Matches(msg, Keys.Bookmarks.Load):
			if v.isSaveRow() {
				v.naming = true
				v.nameInput.SetValue("")
				v.nameInput.Focus()
				return v, nil
			}
			queries := v.parent.config.GetSavedQueries()
			idx := v.queryIndex()
			if idx >= 0 && idx < len(queries) {
				v.parent.editorView.textarea.SetValue(queries[idx].Query)
				if !v.parent.PopView() {
					v.parent.mode = ViewEditor
				}
				return v, v.parent.SetStatus("Loaded: " + queries[idx].Name)
			}
			return v, nil
		case key.Matches(msg, Keys.Bookmarks.Save):
			if v.editorQuery != "" {
				v.naming = true
				v.nameInput.SetValue("")
				v.nameInput.Focus()
				return v, nil
			}
			return v, nil
		case key.Matches(msg, Keys.Bookmarks.Delete):
			if v.isSaveRow() {
				return v, nil
			}
			queries := v.parent.config.GetSavedQueries()
			idx := v.queryIndex()
			if idx >= 0 && idx < len(queries) {
				name := queries[idx].Name
				v.parent.config.DeleteSavedQuery(name)
				if v.cursor >= v.itemCount() && v.cursor > 0 {
					v.cursor--
				}
				return v, tea.Batch(v.parent.requestConfigSave(), v.parent.SetStatus("Deleted: "+name))
			}
			return v, nil
		}
	}
	return v, nil
}
func (v *BookmarksView) View() string {
	var b strings.Builder
	b.WriteString(styles.RenderTitle("Bookmarks"))
	b.WriteString("\n\n")
	if v.naming {
		b.WriteString("  Bookmark name:\n")
		b.WriteString("  " + v.nameInput.View())
		b.WriteString("\n\n")
		b.WriteString(styles.RenderMuted("  Press Enter to save, Esc to cancel"))
		b.WriteString("\n\n")
		b.WriteString(RenderBindingHelpWidth(v.width, Keys.Global.Back))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	queries := v.parent.config.GetSavedQueries()
	row := 0
	if v.editorQuery != "" {
		prefix := "  "
		if v.cursor == row {
			prefix = styles.RenderKey("> ")
		}
		label := "+ Save current query"
		if v.cursor == row {
			b.WriteString(prefix + styles.ActiveListItemStyle.Render(label))
		} else {
			b.WriteString(prefix + styles.RenderKey(label))
		}
		b.WriteString("\n\n")
		row++
	}
	if len(queries) == 0 {
		b.WriteString(styles.RenderMuted("No saved bookmarks"))
	} else {
		for i, sq := range queries {
			prefix := "  "
			if v.cursor == row+i {
				prefix = styles.RenderKey("> ")
			}
			preview := strings.ReplaceAll(sq.Query, "\n", " ")
			maxPreview := 50
			if v.width > 30 {
				maxPreview = v.width - 30
			}
			if len(preview) > maxPreview {
				preview = preview[:maxPreview] + "..."
			}
			nameStr := sq.Name
			line := nameStr + "  " + styles.MutedStyle.Render(preview)
			if v.cursor == row+i {
				b.WriteString(prefix + styles.ActiveListItemStyle.Render(nameStr) + "  " + styles.MutedStyle.Render(preview))
			} else {
				b.WriteString(prefix + line)
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n\n")
	bindings := []key.Binding{
		Keys.Bookmarks.Up,
		Keys.Bookmarks.Down,
		Keys.Bookmarks.Load,
		Keys.Bookmarks.Delete,
	}
	if v.editorQuery != "" {
		bindings = append(bindings, Keys.Bookmarks.Save)
	}
	bindings = append(bindings, Keys.Global.Back, Keys.Global.Quit)
	b.WriteString(RenderBindingHelpWidth(v.width, bindings...))
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}