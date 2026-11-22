package tui
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clidey/whodb/cli/pkg/styles"
	"github.com/clidey/whodb/core/src/engine"
)
type ExportView struct {
	parent           *MainModel
	width            int
	height           int
	filenameInput    textinput.Model
	selectedFormat   int
	selectedDelim    int
	focusIndex       int
	schema           string
	tableName        string
	exporting        bool
	exportSuccess    bool
	exportError      error
	queryResults     *engine.GetRowsResult
	isQueryExport    bool
	savedFilePath    string
	overwrite        bool
	confirmOverwrite bool
	confirmIndex     int
	pendingPath      string
}
var exportFormats = []string{"CSV", "Excel"}
var exportDelimiters = []string{",", ";", "|"}
var exportDelimiterLabels = []string{"Comma (,)", "Semicolon (;)", "Pipe (|)"}
func NewExportView(parent *MainModel) *ExportView {
	ti := textinput.New()
	ti.Placeholder = "export"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.Foreground)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(styles.Primary)
	return &ExportView{
		parent:           parent,
		filenameInput:    ti,
		selectedFormat:   0,
		selectedDelim:    0,
		focusIndex:       0,
		overwrite:        false,
		confirmOverwrite: false,
		confirmIndex:     0,
		pendingPath:      "",
	}
}
func (v *ExportView) SetExportData(schema, tableName string) {
	v.schema = schema
	v.tableName = tableName
	v.filenameInput.SetValue(tableName)
	v.exporting = false
	v.exportSuccess = false
	v.exportError = nil
	v.queryResults = nil
	v.isQueryExport = false
	v.savedFilePath = ""
	v.overwrite = false
	v.confirmOverwrite = false
	v.confirmIndex = 0
	v.pendingPath = ""
}
func (v *ExportView) SetExportDataFromQuery(results *engine.GetRowsResult) {
	v.queryResults = results
	v.isQueryExport = true
	v.schema = ""
	v.tableName = ""
	v.filenameInput.SetValue("query_results")
	v.exporting = false
	v.exportSuccess = false
	v.exportError = nil
	v.savedFilePath = ""
	v.overwrite = false
	v.confirmOverwrite = false
	v.confirmIndex = 0
	v.pendingPath = ""
}
func (v *ExportView) Update(msg tea.Msg) (*ExportView, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.filenameInput.Width = clamp(msg.Width-12, 20, 60)
		return v, nil
	case exportResultMsg:
		v.exporting = false
		if msg.success {
			v.exportSuccess = true
			v.exportError = nil
			v.savedFilePath = msg.savedFilePath
			v.confirmOverwrite = false
			v.confirmIndex = 0
			v.pendingPath = ""
			return v, v.parent.SetStatus("Export complete")
		}
		v.exportSuccess = false
		v.exportError = msg.err
		return v, nil
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			v.focusIndex--
			if v.focusIndex < 0 {
				v.focusIndex = v.maxIndex()
			}
			if v.focusIndex == 0 {
				v.filenameInput.Focus()
			} else {
				v.filenameInput.Blur()
			}
			return v, nil
		case tea.MouseButtonWheelDown:
			v.focusIndex++
			if v.focusIndex > v.maxIndex() {
				v.focusIndex = 0
			}
			if v.focusIndex == 0 {
				v.filenameInput.Focus()
			} else {
				v.filenameInput.Blur()
			}
			return v, nil
		}
	case tea.KeyMsg:
		if v.confirmOverwrite {
			switch msg.String() {
			case "esc":
				v.confirmOverwrite = false
				v.confirmIndex = 0
				return v, nil
			case "left", "h", "right", "l", "tab", "down", "up", "j", "k":
				if v.confirmIndex == 0 {
					v.confirmIndex = 1
				} else {
					v.confirmIndex = 0
				}
				return v, nil
			case "enter":
				if v.confirmIndex == 0 {
					return v, v.performExport()
				}
				v.confirmOverwrite = false
				v.confirmIndex = 0
				return v, nil
			}
			return v, nil
		}
		switch {
		case key.Matches(msg, Keys.Global.Back):
			if !v.exporting {
				if !v.parent.PopView() {
					v.parent.mode = ViewResults
				}
				return v, nil
			}
		case key.Matches(msg, Keys.Export.Export):
			expIdx := v.exportButtonIndex()
			cancelIdx := v.cancelButtonIndex()
			owIdx := v.overwriteIndex()
			if v.focusIndex == expIdx {
				input := v.filenameInput.Value()
				format := exportFormats[v.selectedFormat]
				resolved, willOverwrite, err := resolveExportPath(input, format, v.overwrite)
				if err != nil {
					v.exportError = err
					return v, nil
				}
				v.pendingPath = resolved
				if willOverwrite {
					v.confirmOverwrite = true
					v.confirmIndex = 0
					return v, nil
				}
				return v, v.performExport()
			} else if v.focusIndex == cancelIdx {
				if !v.parent.PopView() {
					v.parent.mode = ViewResults
				}
				return v, nil
			} else if v.focusIndex == owIdx {
				v.overwrite = !v.overwrite
				return v, nil
			}
		case key.Matches(msg, Keys.Export.Next):
			v.focusIndex++
			if v.focusIndex > v.maxIndex() {
				v.focusIndex = 0
			}
			if v.focusIndex == 0 {
				v.filenameInput.Focus()
			} else {
				v.filenameInput.Blur()
			}
			return v, nil
		case key.Matches(msg, Keys.Export.Prev):
			v.focusIndex--
			if v.focusIndex < 0 {
				v.focusIndex = v.maxIndex()
			}
			if v.focusIndex == 0 {
				v.filenameInput.Focus()
			} else {
				v.filenameInput.Blur()
			}
			return v, nil
		case key.Matches(msg, Keys.Export.OptionLeft):
			if v.focusIndex == 1 {
				v.selectedFormat--
				if v.selectedFormat < 0 {
					v.selectedFormat = len(exportFormats) - 1
				}
				return v, nil
			} else if v.hasDelimiter() && v.focusIndex == 2 {
				v.selectedDelim--
				if v.selectedDelim < 0 {
					v.selectedDelim = len(exportDelimiters) - 1
				}
				return v, nil
			} else if v.focusIndex == v.overwriteIndex() {
				v.overwrite = !v.overwrite
				return v, nil
			}
		case key.Matches(msg, Keys.Export.OptionRight):
			if v.focusIndex == 1 {
				v.selectedFormat++
				if v.selectedFormat >= len(exportFormats) {
					v.selectedFormat = 0
				}
				return v, nil
			} else if v.hasDelimiter() && v.focusIndex == 2 {
				v.selectedDelim++
				if v.selectedDelim >= len(exportDelimiters) {
					v.selectedDelim = 0
				}
				return v, nil
			} else if v.focusIndex == v.overwriteIndex() {
				v.overwrite = !v.overwrite
				return v, nil
			}
		}
	}
	if v.focusIndex == 0 {
		v.filenameInput, cmd = v.filenameInput.Update(msg)
	}
	return v, cmd
}
func (v *ExportView) View() string {
	var b strings.Builder
	b.WriteString(styles.RenderTitle("Export Data"))
	b.WriteString("\n\n")
	if v.exporting {
		b.WriteString(v.parent.SpinnerView() + styles.RenderMuted(" Exporting..."))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	if v.exportSuccess {
		b.WriteString(styles.RenderOk("Export completed successfully!"))
		b.WriteString("\n\n")
		if v.savedFilePath != "" {
			b.WriteString(styles.RenderMuted(fmt.Sprintf("File saved to: %s", v.savedFilePath)))
			b.WriteString("\n\n")
		}
		b.WriteString(styles.RenderMuted("Press ESC to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	if v.exportError != nil {
		b.WriteString(styles.RenderErrorBox(fmt.Sprintf("Export failed: %s", v.exportError.Error())))
		b.WriteString("\n\n")
		b.WriteString(styles.RenderMuted("Press ESC to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	if v.confirmOverwrite {
		b.WriteString(styles.RenderTitle("Confirm Overwrite"))
		b.WriteString("\n\n")
		path := v.pendingPath
		if path == "" {
			input := v.filenameInput.Value()
			format := exportFormats[v.selectedFormat]
			if p, _, err := resolveExportPath(input, format, true); err == nil {
				path = p
			}
		}
		b.WriteString(styles.RenderMuted("File exists:"))
		b.WriteString("\n  ")
		b.WriteString(styles.RenderKey(path))
		b.WriteString("\n\n")
		b.WriteString(styles.RenderMuted("Overwrite this file?"))
		b.WriteString("\n\n  ")
		if v.confirmIndex == 0 {
			b.WriteString(styles.ActiveListItemStyle.Render(" Yes "))
			b.WriteString("  ")
			b.WriteString(styles.RenderMuted("[No]"))
		} else {
			b.WriteString(styles.RenderKey("[Yes]"))
			b.WriteString("  ")
			b.WriteString(styles.ActiveListItemStyle.Render(" No "))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.RenderHelp(
			"←/→", "select",
			"enter", "confirm",
			"esc", "cancel",
		))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	if v.isQueryExport {
		b.WriteString(styles.RenderMuted("Exporting query results"))
	} else {
		b.WriteString(styles.RenderMuted(fmt.Sprintf("Exporting table: %s.%s", v.schema, v.tableName)))
	}
	b.WriteString("\n\n")
	filenameLabel := "Filename:"
	if v.focusIndex == 0 {
		filenameLabel = styles.RenderKey("▶ " + filenameLabel)
	} else {
		filenameLabel = "  " + filenameLabel
	}
	b.WriteString(filenameLabel)
	b.WriteString("\n")
	b.WriteString("  " + v.filenameInput.View())
	b.WriteString("\n")
	filename := v.filenameInput.Value()
	if filename == "" {
		filename = v.filenameInput.Placeholder
	}
	format := "CSV"
	if v.selectedFormat >= 0 && v.selectedFormat < len(exportFormats) {
		format = exportFormats[v.selectedFormat]
	}
	if candidate, willOverwrite, err := resolveExportPath(filename, format, v.overwrite); err == nil {
		if _, origExists, _ := resolveExportPath(filename, format, true); true {
			b.WriteString("  ")
			if willOverwrite {
				b.WriteString(styles.RenderKey(fmt.Sprintf("Will overwrite: %s", candidate)))
			} else {
				b.WriteString(styles.RenderMuted(fmt.Sprintf("Will save to: %s", candidate)))
				if !v.overwrite && origExists {
					b.WriteString("\n  ")
					b.WriteString(styles.RenderMuted("Existing file detected; using next available name."))
				}
			}
		}
	}
	b.WriteString("\n\n")
	formatLabel := "Format:"
	if v.focusIndex == 1 {
		formatLabel = styles.RenderKey("▶ " + formatLabel)
	} else {
		formatLabel = "  " + formatLabel
	}
	b.WriteString(formatLabel)
	b.WriteString("\n")
	b.WriteString("  ")
	for i, format := range exportFormats {
		if i == v.selectedFormat {
			if v.focusIndex == 1 {
				b.WriteString(styles.ActiveListItemStyle.Render(fmt.Sprintf(" %s ", format)))
			} else {
				b.WriteString(styles.RenderKey(fmt.Sprintf("[%s]", format)))
			}
		} else {
			b.WriteString(styles.RenderMuted(fmt.Sprintf(" %s ", format)))
		}
		if i < len(exportFormats)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n\n")
	if v.hasDelimiter() {
		delimLabel := "Delimiter:"
		if v.focusIndex == 2 {
			delimLabel = styles.RenderKey("▶ " + delimLabel)
		} else {
			delimLabel = "  " + delimLabel
		}
		b.WriteString(delimLabel)
		b.WriteString("\n")
		b.WriteString("  ")
		for i, label := range exportDelimiterLabels {
			if i == v.selectedDelim {
				if v.focusIndex == 2 {
					b.WriteString(styles.ActiveListItemStyle.Render(fmt.Sprintf(" %s ", label)))
				} else {
					b.WriteString(styles.RenderKey(fmt.Sprintf("[%s]", label)))
				}
			} else {
				b.WriteString(styles.RenderMuted(fmt.Sprintf(" %s ", label)))
			}
			if i < len(exportDelimiterLabels)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n")
	}
	owLabel := "Overwrite existing:"
	if v.focusIndex == v.overwriteIndex() {
		owLabel = styles.RenderKey("▶ " + owLabel)
	} else {
		owLabel = "  " + owLabel
	}
	b.WriteString(owLabel)
	b.WriteString("\n  ")
	if v.overwrite {
		b.WriteString(styles.ActiveListItemStyle.Render(" On "))
		b.WriteString("  ")
		b.WriteString(styles.RenderMuted("[Off]"))
	} else {
		b.WriteString(styles.RenderKey("[On]"))
		b.WriteString("  ")
		b.WriteString(styles.ActiveListItemStyle.Render(" Off "))
	}
	b.WriteString("\n\n")
	b.WriteString("  ")
	if v.focusIndex == v.exportButtonIndex() {
		b.WriteString(styles.ActiveListItemStyle.Render(" Export "))
	} else {
		b.WriteString(styles.RenderKey("[Export]"))
	}
	b.WriteString("  ")
	if v.focusIndex == v.cancelButtonIndex() {
		b.WriteString(styles.ActiveListItemStyle.Render(" Cancel "))
	} else {
		b.WriteString(styles.RenderMuted("[Cancel]"))
	}
	b.WriteString("\n\n")
	b.WriteString(renderBindingHelpWidthNoHelp(v.width,
		Keys.Export.Prev,
		Keys.Export.Next,
		Keys.Export.OptionLeft,
		Keys.Export.OptionRight,
		Keys.Export.Export,
		Keys.Global.Back,
	))
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
func (v *ExportView) performExport() tea.Cmd {
	v.exporting = true
	filename := v.filenameInput.Value()
	format := exportFormats[v.selectedFormat]
	delimiter := exportDelimiters[v.selectedDelim]
	pendingPath := v.pendingPath
	overwrite := v.overwrite
	isQueryExport := v.isQueryExport
	queryResults := v.queryResults
	schema := v.schema
	tableName := v.tableName
	dbManager := v.parent.dbManager
	return func() tea.Msg {
		var fullFilename string
		var err error
		if pendingPath != "" {
			fullFilename = pendingPath
		} else {
			fullFilename, _, err = resolveExportPath(filename, format, overwrite)
			if err != nil {
				return exportResultMsg{success: false, err: err}
			}
		}
		dir := filepath.Dir(fullFilename)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return exportResultMsg{success: false, err: fmt.Errorf("failed to create directory: %w", err)}
			}
		}
		if isQueryExport {
			if format == "CSV" {
				err = dbManager.ExportResultsToCSV(queryResults, fullFilename, delimiter)
			} else {
				err = dbManager.ExportResultsToExcel(queryResults, fullFilename)
			}
		} else {
			if format == "CSV" {
				err = dbManager.ExportToCSV(schema, tableName, fullFilename, delimiter)
			} else {
				err = dbManager.ExportToExcel(schema, tableName, fullFilename)
			}
		}
		if err != nil {
			return exportResultMsg{success: false, err: err}
		}
		return exportResultMsg{success: true, err: nil, savedFilePath: fullFilename}
	}
}
func resolveExportPath(input, format string, overwrite bool) (string, bool, error) {
	s := strings.ReplaceAll(strings.TrimSpace(input), "\x00", "")
	if s == "" {
		return "", false, fmt.Errorf("invalid path or filename")
	}
	if strings.HasPrefix(s, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			if s == "~" {
				s = home
			} else if strings.HasPrefix(s, "~/") || strings.HasPrefix(s, "~\\") {
				s = filepath.Join(home, s[2:])
			}
		}
	}
	p := filepath.Clean(s)
	if !filepath.IsAbs(p) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false, fmt.Errorf("failed to get working directory: %w", err)
		}
		p = filepath.Join(cwd, p)
	}
	wantExt := ".csv"
	if strings.EqualFold(format, "Excel") {
		wantExt = ".xlsx"
	}
	curExt := strings.ToLower(filepath.Ext(p))
	if curExt == "" {
		p = p + wantExt
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p
	}
	if info, err := os.Stat(abs); err == nil && !info.IsDir() {
		if overwrite {
			return abs, true, nil
		}
		dir := filepath.Dir(abs)
		base := filepath.Base(abs)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		for i := 1; ; i++ {
			candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, i, ext))
			if _, statErr := os.Stat(candidate); os.IsNotExist(statErr) {
				return candidate, false, nil
			}
		}
	}
	return abs, false, nil
}
func (v *ExportView) hasDelimiter() bool {
	return v.selectedFormat >= 0 && v.selectedFormat < len(exportFormats) && exportFormats[v.selectedFormat] == "CSV"
}
func (v *ExportView) overwriteIndex() int {
	if v.hasDelimiter() {
		return 3
	}
	return 2
}
func (v *ExportView) exportButtonIndex() int {
	return v.overwriteIndex() + 1
}
func (v *ExportView) cancelButtonIndex() int {
	return v.overwriteIndex() + 2
}
func (v *ExportView) maxIndex() int {
	return v.cancelButtonIndex()
}