package tui
import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clidey/whodb/cli/internal/history"
	"github.com/clidey/whodb/cli/pkg/styles"
)
type historyItem struct {
	entry history.Entry
}
func (i historyItem) Title() string {
	query := i.entry.Query
	if len(query) > 60 {
		query = query[:60] + "..."
	}
	return query
}
func (i historyItem) Description() string {
	status := "✓"
	if !i.entry.Success {
		status = "✗"
	}
	return fmt.Sprintf("%s %s - %s", status, i.entry.Database, i.entry.Timestamp.Format("2006-01-02 15:04:05"))
}
func (i historyItem) FilterValue() string {
	return i.entry.Query
}
type HistoryView struct {
	parent          *MainModel
	list            list.Model
	width           int
	height          int
	confirmingClear bool
	executing       bool
	queryCancel     context.CancelFunc
	retryPrompt     RetryPrompt
}
func NewHistoryView(parent *MainModel) *HistoryView {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Query History"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	return &HistoryView{
		parent: parent,
		list:   l,
	}
}
func (v *HistoryView) Update(msg tea.Msg) (*HistoryView, tea.Cmd) {
	switch msg := msg.(type) {
	case HistoryQueryMsg:
		v.executing = false
		v.queryCancel = nil
		if msg.Err != nil {
			if errors.Is(msg.Err, context.DeadlineExceeded) {
				v.parent.err = fmt.Errorf("query timed out")
				v.retryPrompt.Show(msg.Query)
			} else if errors.Is(msg.Err, context.Canceled) {
				return v, nil
			} else {
				v.parent.err = msg.Err
			}
			return v, nil
		}
		v.parent.resultsView.SetResults(msg.Result, msg.Query)
		v.parent.PushView(ViewResults)
		return v, nil
	case tea.WindowSizeMsg:
		overhead := 10
		h := msg.Height - overhead
		if h < 5 {
			h = 5
		}
		v.list.SetSize(msg.Width-4, h)
		return v, nil
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			v.list.CursorUp()
			return v, nil
		case tea.MouseButtonWheelDown:
			v.list.CursorDown()
			return v, nil
		}
	case tea.KeyMsg:
		if v.retryPrompt.IsActive() {
			result, handled := v.retryPrompt.HandleKeyMsg(msg.String())
			if handled {
				if result != nil {
					v.parent.err = nil
					return v, v.executeQueryWithTimeout(v.retryPrompt.TimedOutQuery(), result.Timeout)
				}
				return v, nil
			}
			return v, nil
		}
		switch {
		case key.Matches(msg, Keys.History.Edit):
			if item, ok := v.list.SelectedItem().(historyItem); ok {
				v.parent.editorView.textarea.SetValue(item.entry.Query)
				v.parent.PushView(ViewEditor)
				return v, nil
			}
		case key.Matches(msg, Keys.History.Rerun):
			if v.executing {
				return v, nil
			}
			if item, ok := v.list.SelectedItem().(historyItem); ok {
				v.executing = true
				v.retryPrompt.SetAutoRetried(false)
				query := item.entry.Query
				timeout := v.parent.config.GetQueryTimeout()
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				v.queryCancel = cancel
				return v, func() tea.Msg {
					defer cancel()
					result, err := v.parent.dbManager.ExecuteQueryWithContext(ctx, query)
					return HistoryQueryMsg{Result: result, Query: query, Err: err}
				}
			}
		case key.Matches(msg, Keys.History.ClearAll):
			if !v.confirmingClear {
				v.confirmingClear = true
				return v, nil
			}
		case msg.String() == "y" || msg.String() == "Y":
			if v.confirmingClear {
				v.parent.histMgr.Clear()
				v.refreshList()
				v.confirmingClear = false
				return v, nil
			}
		case msg.String() == "n" || msg.String() == "N":
			if v.confirmingClear {
				v.confirmingClear = false
				return v, nil
			}
		case key.Matches(msg, Keys.Global.Back):
			if v.executing && v.queryCancel != nil {
				v.queryCancel()
				return v, nil
			}
			if v.confirmingClear {
				v.confirmingClear = false
				return v, nil
			}
			if !v.parent.PopView() {
				v.parent.mode = ViewBrowser
			}
			return v, nil
		}
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}
func (v *HistoryView) View() string {
	var b strings.Builder
	b.WriteString(styles.RenderTitle("Query History"))
	b.WriteString("\n\n")
	if v.executing {
		b.WriteString(v.parent.SpinnerView() + styles.RenderMuted(" Executing query... Press ESC to cancel"))
		b.WriteString("\n\n")
	}
	if v.retryPrompt.IsActive() {
		b.WriteString(v.retryPrompt.View())
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	if v.confirmingClear {
		b.WriteString(styles.RenderErr("⚠ Clear all history?"))
		b.WriteString("\n\n")
		b.WriteString(styles.RenderMuted("This will delete all query history entries."))
		b.WriteString("\n\n")
		b.WriteString(styles.RenderHelp(
			"[y]", "confirm",
			"[n]/esc", "cancel",
		))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	entries := v.parent.histMgr.GetAll()
	if len(entries) == 0 {
		b.WriteString(styles.RenderMuted("No history entries"))
	} else {
		b.WriteString(v.list.View())
	}
	b.WriteString("\n\n")
	b.WriteString(RenderBindingHelpWidth(v.width,
		Keys.Browser.Up,
		Keys.Browser.Down,
		Keys.History.Edit,
		Keys.History.Rerun,
		Keys.History.ClearAll,
		Keys.Global.NextView,
		Keys.Global.Back,
		Keys.Global.Quit,
	))
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
func (v *HistoryView) refreshList() {
	entries := v.parent.histMgr.GetAll()
	items := make([]list.Item, len(entries))
	for i, entry := range entries {
		items[i] = historyItem{entry: entry}
	}
	v.list.SetItems(items)
}
func (v *HistoryView) Init() {
	v.refreshList()
}
func (v *HistoryView) executeQueryWithTimeout(query string, timeout time.Duration) tea.Cmd {
	v.executing = true
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	v.queryCancel = cancel
	return func() tea.Msg {
		defer cancel()
		result, err := v.parent.dbManager.ExecuteQueryWithContext(ctx, query)
		return HistoryQueryMsg{Result: result, Query: query, Err: err}
	}
}