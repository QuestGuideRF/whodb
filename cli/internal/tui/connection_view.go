package tui
import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	cloudruntime "github.com/clidey/whodb/cli/internal/cloud"
	"github.com/clidey/whodb/cli/internal/config"
	"github.com/clidey/whodb/cli/internal/connectionopts"
	dbmgr "github.com/clidey/whodb/cli/internal/database"
	"github.com/clidey/whodb/cli/internal/docker"
	"github.com/clidey/whodb/cli/internal/sourcetypes"
	"github.com/clidey/whodb/cli/pkg/styles"
)
type connectionItem struct {
	conn        config.Connection
	source      string
	description string
}
func (i connectionItem) Title() string { return i.conn.Name }
const ConnectionSourceDocker = "docker"
const ConnectionSourceCloud = "cloud"
func (i connectionItem) Description() string {
	if i.description != "" {
		return i.description
	}
	desc := fmt.Sprintf("%s@%s", i.conn.Type, i.conn.Host)
	if i.source == dbmgr.ConnectionSourceEnv {
		desc += " (env)"
	} else if i.source == ConnectionSourceDocker {
		desc += " (docker)"
	} else if i.source == ConnectionSourceCloud {
		desc += " (cloud)"
	}
	return desc
}
func (i connectionItem) FilterValue() string { return i.conn.Name }
type connectionPingResult struct {
	checked bool
	online  bool
}
type connectionPingMsg struct {
	name   string
	online bool
}
func buildConnectionStatusMessage(conn *config.Connection, sslSummary string) string {
	parts := make([]string, 0, 2)
	if conn != nil {
		parts = append(parts, fmt.Sprintf("Connected to %s@%s", conn.Type, conn.Host))
	}
	if sslSummary != "" {
		parts = append(parts, sslSummary)
	}
	return strings.Join(parts, " • ")
}
type connectionDelegate struct {
	pingResults map[string]connectionPingResult
}
func (d connectionDelegate) Height() int                             { return 2 }
func (d connectionDelegate) Spacing() int                            { return 1 }
func (d connectionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d connectionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(connectionItem)
	if !ok {
		return
	}
	str := ""
	if index == m.Index() {
		str = styles.ActiveListItemStyle.Render("▶ " + i.Title())
	} else {
		str = "  " + i.Title()
	}
	desc := styles.RenderMuted(i.Description())
	if result, ok := d.pingResults[i.conn.Name]; ok && result.checked {
		if result.online {
			desc += " " + styles.SuccessStyle.Render("●")
		} else {
			desc += " " + styles.ErrorStyle.Render("●")
		}
	} else if i.conn.SSHHost != "" {
		desc += " " + styles.MutedStyle.Render("○")
	}
	str += "\n  " + desc
	fmt.Fprint(w, str)
}
const (
	fieldName        = 0
	fieldHost        = 1
	fieldPort        = 2
	fieldUsername    = 3
	fieldPassword    = 4
	fieldDatabase    = 5
	fieldSchema      = 6
	fieldSSHHost     = 7
	fieldSSHUser     = 8
	fieldSSHKeyFile  = 9
	fieldSSHPassword = 10
	fieldSSLCAFile   = 11
	fieldSSLCertFile = 12
	fieldSSLKeyFile  = 13
	fieldSSLServer   = 14
)
const (
	focusDBType    = 15
	focusSSHToggle = 16
	focusSSLMode   = 17
	focusConnect   = 18
)
type ConnectionView struct {
	parent        *MainModel
	list          list.Model
	mode          string
	inputs        []textinput.Model
	focusIndex    int
	dbTypes       []string
	dbTypeIndex   int
	visibleFields []int
	sshEnabled    bool
	sslModeIndex  int
	connecting    bool
	connError     error
	awaitingPassword bool
	passwordPrompt   textinput.Model
	escPressed     bool
	escTimeoutSecs int
	formViewport viewport.Model
	formReady    bool
	width        int
	height       int
	pingResults   map[string]connectionPingResult
	dockerItems   []connectionItem
	cloudItems    []connectionItem
	pingQueue     []config.Connection
	advanced      map[string]string
	selectorCache map[string]wrappedSelectableOptionsCacheEntry
}
type wrappedSelectableOptionsCacheEntry struct {
	rendered  string
	lineCount int
}
func NewConnectionView(parent *MainModel) *ConnectionView {
	items := baseConnectionItems(parent.dbManager.ListConnectionsWithSource())
	pingResults := make(map[string]connectionPingResult)
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	l := list.New(listItems, connectionDelegate{pingResults: pingResults}, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetStatusBarItemName("connection", "connections")
	newInput := func(placeholder string, charLimit int) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.CharLimit = charLimit
		ti.Width = 40
		ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
		ti.TextStyle = lipgloss.NewStyle().Foreground(styles.Foreground)
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(styles.Primary)
		return ti
	}
	inputs := make([]textinput.Model, 15)
	inputs[fieldName] = newInput("My Connection", 50)
	inputs[fieldHost] = newInput("localhost", 100)
	inputs[fieldPort] = newInput("", 5)
	inputs[fieldUsername] = newInput("postgres", 50)
	inputs[fieldPassword] = newInput("password", 100)
	inputs[fieldPassword].EchoMode = textinput.EchoPassword
	inputs[fieldPassword].EchoCharacter = '•'
	inputs[fieldDatabase] = newInput("mydb", 50)
	inputs[fieldSchema] = newInput("Schema name (optional)", 50)
	inputs[fieldSSHHost] = newInput("ssh.example.com", 100)
	inputs[fieldSSHUser] = newInput("ssh-user", 50)
	inputs[fieldSSHKeyFile] = newInput("~/.ssh/id_rsa", 200)
	inputs[fieldSSHPassword] = newInput("SSH password (optional)", 100)
	inputs[fieldSSHPassword].EchoMode = textinput.EchoPassword
	inputs[fieldSSHPassword].EchoCharacter = '•'
	inputs[fieldSSLCAFile] = newInput("./ca.pem", 200)
	inputs[fieldSSLCertFile] = newInput("./client-cert.pem", 200)
	inputs[fieldSSLKeyFile] = newInput("./client-key.pem", 200)
	inputs[fieldSSLServer] = newInput("db.internal", 100)
	mode := "list"
	fi := focusDBType
	if len(items) == 0 {
		mode = "form"
	}
	prompt := newInput("enter password", 100)
	prompt.EchoMode = textinput.EchoPassword
	prompt.EchoCharacter = '•'
	dbTypes := sourcetypes.IDs()
	view := &ConnectionView{
		parent:           parent,
		list:             l,
		mode:             mode,
		inputs:           inputs,
		focusIndex:       fi,
		dbTypes:          dbTypes,
		dbTypeIndex:      0,
		visibleFields:    getVisibleFields(dbTypes[0]),
		sslModeIndex:     0,
		connecting:       false,
		awaitingPassword: false,
		passwordPrompt:   prompt,
		pingResults:      pingResults,
		selectorCache:    make(map[string]wrappedSelectableOptionsCacheEntry),
	}
	view.updatePortPlaceholder()
	return view
}
func (v *ConnectionView) Update(msg tea.Msg) (*ConnectionView, tea.Cmd) {
	if v.mode == "form" {
		return v.updateForm(msg)
	}
	return v.updateList(msg)
}
func (v *ConnectionView) Init() tea.Cmd {
	cmds := []tea.Cmd{v.loadDockerConnections(), v.loadCloudConnections()}
	if v.mode == "list" && len(v.list.Items()) > 0 {
		cmds = append(cmds, v.pingAllConnections())
	}
	return tea.Batch(cmds...)
}
func (v *ConnectionView) pingAllConnections() tea.Cmd {
	v.pingQueue = nil
	for _, item := range v.list.Items() {
		ci, ok := item.(connectionItem)
		if !ok {
			continue
		}
		if ci.conn.SSHHost != "" || ci.source == ConnectionSourceDocker || ci.source == ConnectionSourceCloud {
			continue
		}
		v.pingQueue = append(v.pingQueue, ci.conn)
	}
	return v.pingNextConnection()
}
func (v *ConnectionView) updateList(msg tea.Msg) (*ConnectionView, tea.Cmd) {
	switch msg := msg.(type) {
	case dockerConnectionsLoadedMsg:
		v.dockerItems = msg.items
		v.setListItems(v.selectedConnectionName())
		return v, nil
	case cloudConnectionsLoadedMsg:
		v.cloudItems = msg.items
		v.setListItems(v.selectedConnectionName())
		return v, nil
	case connectionPingMsg:
		v.pingResults[msg.name] = connectionPingResult{checked: true, online: msg.online}
		return v, v.pingNextConnection()
	case connectionResultMsg:
		v.connecting = false
		if msg.err != nil {
			v.parent.err = msg.err
			return v, nil
		}
		v.parent.currentProfileName = ""
		v.parent.mode = ViewBrowser
		v.parent.initLayout()
		conn := v.parent.dbManager.GetCurrentConnection()
		connDesc := msg.statusMessage
		if connDesc == "" {
			connDesc = buildConnectionStatusMessage(conn, "")
		}
		return v, tea.Batch(v.parent.browserView.Init(), v.parent.SetStatus(connDesc))
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
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
	case escTimeoutTickMsg:
		if v.escPressed {
			v.escTimeoutSecs--
			if v.escTimeoutSecs <= 0 {
				v.escPressed = false
				v.escTimeoutSecs = 0
			}
			return v, tea.Tick(time.Second, func(time.Time) tea.Msg { return escTimeoutTickMsg{} })
		}
		return v, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.ConnectionList.Down):
			v.list.CursorDown()
			return v, nil
		case key.Matches(msg, Keys.ConnectionList.Up):
			v.list.CursorUp()
			return v, nil
		case key.Matches(msg, Keys.ConnectionList.Connect):
			if item, ok := v.list.SelectedItem().(connectionItem); ok {
				if item.source == ConnectionSourceDocker || item.source == ConnectionSourceCloud {
					v.mode = "form"
					v.resetForm()
					v.prefillFromConnection(item.conn)
					return v, nil
				}
				v.connecting = true
				v.connError = nil
				conn := item.conn
				return v, func() tea.Msg {
					if err := v.parent.dbManager.Connect(&conn); err != nil {
						return connectionResultMsg{err: err}
					}
					sslSummary, err := v.parent.dbManager.GetSSLStatusSummary()
					if err != nil {
						sslSummary = ""
					}
					return connectionResultMsg{
						err:           nil,
						statusMessage: buildConnectionStatusMessage(&conn, sslSummary),
					}
				}
			}
		case key.Matches(msg, Keys.ConnectionList.New):
			v.mode = "form"
			v.resetForm()
			v.inputs[0].Focus()
			return v, nil
		case key.Matches(msg, Keys.ConnectionList.DeleteConn):
			if item, ok := v.list.SelectedItem().(connectionItem); ok {
				if item.source != dbmgr.ConnectionSourceSaved {
					return v, nil
				}
				v.parent.config.RemoveConnection(item.conn.Name)
				v.refreshList()
				return v, tea.Batch(v.parent.requestConfigSave(), v.pingAllConnections(), v.loadDockerConnections(), v.loadCloudConnections())
			}
		case key.Matches(msg, Keys.ConnectionList.QuitEsc):
			if v.escPressed {
				return v, tea.Quit
			}
			v.escPressed = true
			v.escTimeoutSecs = 3
			return v, tea.Tick(time.Second, func(time.Time) tea.Msg { return escTimeoutTickMsg{} })
		}
	}
	if km, ok := msg.(tea.KeyMsg); ok && km.String() != "esc" {
		v.escPressed = false
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}
func (v *ConnectionView) inputWidth() int {
	return clamp(v.width-8, 20, 60)
}
func (v *ConnectionView) updateForm(msg tea.Msg) (*ConnectionView, tea.Cmd) {
	var cmd tea.Cmd
	if pm, ok := msg.(connectionPingMsg); ok {
		v.pingResults[pm.name] = connectionPingResult{checked: true, online: pm.online}
		return v, v.pingNextConnection()
	}
	if dm, ok := msg.(dockerConnectionsLoadedMsg); ok {
		v.dockerItems = dm.items
		if v.mode == "form" && len(v.list.Items()) == 0 && len(dm.items) > 0 && !v.formHasUserInput() {
			v.mode = "list"
			v.setListItems("")
			return v, v.pingAllConnections()
		}
		return v, nil
	}
	if cm, ok := msg.(cloudConnectionsLoadedMsg); ok {
		v.cloudItems = cm.items
		if v.mode == "form" && len(v.list.Items()) == 0 && len(cm.items) > 0 && !v.formHasUserInput() {
			v.mode = "list"
			v.setListItems("")
			return v, v.pingAllConnections()
		}
		return v, nil
	}
	if v.awaitingPassword {
		switch m := msg.(type) {
		case tea.KeyMsg:
			switch m.String() {
			case "enter":
				v.inputs[4].SetValue(v.passwordPrompt.Value())
				v.passwordPrompt.SetValue("")
				v.awaitingPassword = false
				v.connecting = true
				v.connError = nil
				return v, v.connect()
			case "esc":
				v.passwordPrompt.SetValue("")
				v.awaitingPassword = false
				return v, nil
			}
		}
		v.passwordPrompt, cmd = v.passwordPrompt.Update(msg)
		return v, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		if !v.formReady {
			v.formViewport = viewport.New(msg.Width-4, msg.Height-8)
			v.formViewport.MouseWheelEnabled = true
			v.formReady = true
		}
		return v, nil
	case tea.MouseMsg:
		if v.formReady {
			v.formViewport, cmd = v.formViewport.Update(msg)
			return v, cmd
		}
	case connectionResultMsg:
		if msg.err != nil {
			v.connError = msg.err
			v.connecting = false
		} else {
			v.parent.currentProfileName = ""
			v.parent.mode = ViewBrowser
			v.parent.initLayout()
			conn := v.parent.dbManager.GetCurrentConnection()
			connDesc := msg.statusMessage
			if connDesc == "" {
				connDesc = buildConnectionStatusMessage(conn, "")
			}
			return v, tea.Batch(v.parent.browserView.Init(), v.parent.SetStatus(connDesc))
		}
		return v, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return v, tea.Quit
		case "esc":
			if len(v.parent.config.Connections) > 0 {
				v.mode = "list"
				v.connError = nil
				v.refreshList()
				return v, tea.Batch(v.pingAllConnections(), v.loadDockerConnections(), v.loadCloudConnections())
			}
			return v, tea.Quit
		case "tab", "down":
			v.nextInput()
			return v, nil
		case "shift+tab", "up":
			v.prevInput()
			return v, nil
		case "left":
			if v.focusIndex == focusDBType {
				v.dbTypeIndex--
				if v.dbTypeIndex < 0 {
					v.dbTypeIndex = len(v.dbTypes) - 1
				}
				v.onDbTypeChanged()
			} else if v.focusIndex == focusSSLMode {
				sslModes := v.sslModes()
				if len(sslModes) > 0 {
					v.sslModeIndex--
					if v.sslModeIndex < 0 {
						v.sslModeIndex = len(sslModes) - 1
					}
				}
			}
			return v, nil
		case "right":
			if v.focusIndex == focusDBType {
				v.dbTypeIndex++
				if v.dbTypeIndex >= len(v.dbTypes) {
					v.dbTypeIndex = 0
				}
				v.onDbTypeChanged()
			} else if v.focusIndex == focusSSLMode {
				sslModes := v.sslModes()
				if len(sslModes) > 0 {
					v.sslModeIndex++
					if v.sslModeIndex >= len(sslModes) {
						v.sslModeIndex = 0
					}
				}
			}
			return v, nil
		case "enter", " ":
			if v.focusIndex == focusSSHToggle {
				v.sshEnabled = !v.sshEnabled
				v.visibleFields = getVisibleFields(v.dbTypes[v.dbTypeIndex])
				return v, nil
			}
			if v.focusIndex == focusSSLMode {
				if msg.String() == " " {
					sslModes := v.sslModes()
					if len(sslModes) > 0 {
						v.sslModeIndex++
						if v.sslModeIndex >= len(sslModes) {
							v.sslModeIndex = 0
						}
					}
					return v, nil
				}
				v.nextInput()
				return v, nil
			}
			if msg.String() == " " {
				break
			}
			if v.focusIndex == focusConnect {
				if passwordRequired(v.dbTypes[v.dbTypeIndex]) && v.isFieldVisible(fieldPassword) && v.inputs[fieldPassword].Value() == "" {
					v.awaitingPassword = true
					v.passwordPrompt.SetValue("")
					v.passwordPrompt.Focus()
					return v, nil
				}
				v.connecting = true
				v.connError = nil
				return v, v.connect()
			}
			if msg.String() == "enter" {
				v.nextInput()
			}
			return v, nil
		}
	}
	if v.focusIndex >= 0 && v.focusIndex < len(v.inputs) {
		v.inputs[v.focusIndex], cmd = v.inputs[v.focusIndex].Update(msg)
	}
	return v, cmd
}
func (v *ConnectionView) View() string {
	if v.mode == "form" {
		return v.renderForm()
	}
	title := styles.RenderTitle("Welcome to WhoDB!")
	subtitle := styles.RenderMuted("Select an existing connection below, or create a new one with [n]")
	helpText := RenderBindingHelpWidth(v.parent.width,
		Keys.ConnectionList.Up,
		Keys.ConnectionList.Down,
		Keys.ConnectionList.Connect,
		Keys.ConnectionList.New,
		Keys.ConnectionList.DeleteConn,
		Keys.Global.CycleTheme,
		Keys.ConnectionList.QuitEsc,
		Keys.Global.Quit,
	)
	sep := styles.MutedStyle.Render("  ")
	legend := styles.SuccessStyle.Render("●") + styles.RenderMuted(" available") + sep +
		styles.ErrorStyle.Render("●") + styles.RenderMuted(" not available") + sep +
		styles.MutedStyle.Render("○") + styles.RenderMuted(" inactive tunnel")
	chromeHeight := lipgloss.Height(title) + lipgloss.Height(subtitle) + lipgloss.Height(helpText) + 1 + 5
	listHeight := v.parent.ContentHeight() - chromeHeight
	if listHeight < 3 {
		listHeight = 3
	}
	v.list.SetSize(v.parent.width-4, listHeight)
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")
	if v.connecting {
		b.WriteString(v.parent.SpinnerView() + styles.RenderMuted(" Connecting..."))
	} else {
		b.WriteString(v.list.View())
	}
	b.WriteString("\n\n")
	if v.escPressed {
		b.WriteString(styles.RenderErr(fmt.Sprintf("Press ESC again to quit (%ds)", v.escTimeoutSecs)))
	} else {
		b.WriteString(legend)
	}
	b.WriteString("\n")
	b.WriteString(helpText)
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
func (v *ConnectionView) renderForm() string {
	if v.awaitingPassword {
		var b strings.Builder
		b.WriteString(styles.RenderTitle("Enter Password"))
		b.WriteString("\n  ")
		b.WriteString(v.passwordPrompt.View())
		b.WriteString("\n\n")
		b.WriteString(styles.RenderHelp(
			"enter", "confirm",
			"esc", "cancel",
		))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	if v.connecting {
		var cb strings.Builder
		cb.WriteString(styles.RenderTitle("New Database Connection"))
		cb.WriteString(v.parent.SpinnerView() + styles.RenderMuted(" Connecting..."))
		return lipgloss.NewStyle().Padding(1, 2).Render(cb.String())
	}
	iw := v.inputWidth()
	for i := range v.inputs {
		v.inputs[i].Width = iw
	}
	v.passwordPrompt.Width = iw
	var body strings.Builder
	if v.focusIndex == focusDBType {
		body.WriteString("▶ Database Type:")
	} else {
		body.WriteString("  Database Type:")
	}
	body.WriteString("\n  ")
	dbTypeOptions, _ := v.renderWrappedSelectableOptions(v.dbTypes, v.dbTypeIndex, v.focusIndex == focusDBType, v.selectorWrapWidth())
	body.WriteString(dbTypeOptions)
	body.WriteString("\n\n")
	fieldLabels := map[int]string{
		fieldName:     "Connection Name:",
		fieldHost:     "Host:",
		fieldPort:     "Port:",
		fieldUsername: "Username:",
		fieldPassword: "Password:",
		fieldDatabase: "Database:",
		fieldSchema:   "Schema:",
	}
	for _, i := range v.visibleFields {
		label := fieldLabels[i]
		if v.focusIndex == i {
			label = styles.RenderKey("▶ " + label)
		} else {
			label = "  " + label
		}
		body.WriteString(label)
		body.WriteString("\n  ")
		body.WriteString(v.inputs[i].View())
		body.WriteString("\n\n")
	}
	if isNetworkDatabase(v.dbTypes[v.dbTypeIndex]) {
		toggleLabel := "SSH Tunnel:"
		toggleValue := "Off"
		if v.sshEnabled {
			toggleValue = "On"
		}
		if v.focusIndex == focusSSHToggle {
			body.WriteString(styles.RenderKey("▶ " + toggleLabel))
			body.WriteString("  ")
			body.WriteString(styles.ActiveListItemStyle.Render(toggleValue))
		} else {
			body.WriteString("  " + toggleLabel)
			body.WriteString("  ")
			if v.sshEnabled {
				body.WriteString(styles.RenderKey(toggleValue))
			} else {
				body.WriteString(styles.RenderMuted(toggleValue))
			}
		}
		body.WriteString("\n\n")
		if v.sshEnabled {
			sshLabels := map[int]string{
				fieldSSHHost:     "SSH Host:",
				fieldSSHUser:     "SSH User:",
				fieldSSHKeyFile:  "SSH Key File:",
				fieldSSHPassword: "SSH Password:",
			}
			for _, i := range []int{fieldSSHHost, fieldSSHUser, fieldSSHKeyFile, fieldSSHPassword} {
				label := sshLabels[i]
				if v.focusIndex == i {
					label = styles.RenderKey("▶ " + label)
				} else {
					label = "  " + label
				}
				body.WriteString(label)
				body.WriteString("\n  ")
				body.WriteString(v.inputs[i].View())
				body.WriteString("\n\n")
			}
		}
	}
	if sslModes := v.sslModes(); len(sslModes) > 0 {
		if v.focusIndex == focusSSLMode {
			body.WriteString(styles.RenderKey("▶ SSL Mode:"))
		} else {
			body.WriteString("  SSL Mode:")
		}
		body.WriteString("\n  ")
		sslModeOptions, _ := v.renderWrappedSelectableOptions(sslModes, v.sslModeIndex, v.focusIndex == focusSSLMode, v.selectorWrapWidth())
		body.WriteString(sslModeOptions)
		body.WriteString("\n\n")
		if v.sslFieldsVisible() {
			sslLabels := map[int]string{
				fieldSSLCAFile:   "SSL CA File:",
				fieldSSLCertFile: "SSL Client Cert File:",
				fieldSSLKeyFile:  "SSL Client Key File:",
				fieldSSLServer:   "SSL Server Name:",
			}
			for _, i := range []int{fieldSSLCAFile, fieldSSLCertFile, fieldSSLKeyFile, fieldSSLServer} {
				label := sslLabels[i]
				if v.focusIndex == i {
					label = styles.RenderKey("▶ " + label)
				} else {
					label = "  " + label
				}
				body.WriteString(label)
				body.WriteString("\n  ")
				body.WriteString(v.inputs[i].View())
				body.WriteString("\n\n")
			}
		}
	}
	connectBtn := "[Connect]"
	if v.focusIndex == focusConnect {
		connectBtn = styles.ActiveListItemStyle.Render(" Connect ")
	} else {
		connectBtn = styles.RenderKey(connectBtn)
	}
	body.WriteString("  " + connectBtn)
	title := styles.RenderTitle("New Database Connection")
	helpWidth := v.width
	if helpWidth == 0 {
		helpWidth = v.parent.width
	}
	helpText := ""
	if len(v.parent.config.Connections) > 0 {
		helpText = renderFooterHelpPairsWidthNoHelp(helpWidth,
			Keys.ConnectionForm.Navigate.Help().Key, Keys.ConnectionForm.Navigate.Help().Desc,
			Keys.ConnectionForm.TypeLeft.Help().Key, Keys.ConnectionForm.TypeLeft.Help().Desc,
			Keys.ConnectionForm.ConnectForm.Help().Key, Keys.ConnectionForm.ConnectForm.Help().Desc,
			Keys.Global.CycleTheme.Help().Key, Keys.Global.CycleTheme.Help().Desc,
			Keys.Global.Back.Help().Key, Keys.Global.Back.Help().Desc,
			Keys.Global.Quit.Help().Key, Keys.Global.Quit.Help().Desc,
		)
	} else {
		helpText = renderFooterHelpPairsWidthNoHelp(helpWidth,
			Keys.ConnectionForm.Navigate.Help().Key, Keys.ConnectionForm.Navigate.Help().Desc,
			Keys.ConnectionForm.TypeLeft.Help().Key, Keys.ConnectionForm.TypeLeft.Help().Desc,
			Keys.ConnectionForm.ConnectForm.Help().Key, Keys.ConnectionForm.ConnectForm.Help().Desc,
			Keys.Global.CycleTheme.Help().Key, Keys.Global.CycleTheme.Help().Desc,
			Keys.Global.Quit.Help().Key, Keys.Global.Quit.Help().Desc,
		)
	}
	errorBlock := ""
	if v.connError != nil {
		errorBlock = styles.RenderErrorBox(v.connError.Error()) + "\n"
	}
	chromeHeight := lipgloss.Height(title) + lipgloss.Height(errorBlock) + lipgloss.Height(helpText) + 3
	if v.formReady {
		vpHeight := v.parent.ContentHeight() - chromeHeight
		if vpHeight < 3 {
			vpHeight = 3
		}
		v.formViewport.Height = vpHeight
		v.formViewport.Width = v.width - 4
		v.formViewport.SetContent(body.String())
	}
	pad := "  "
	var out strings.Builder
	out.WriteString("\n")
	for _, line := range strings.Split(title, "\n") {
		out.WriteString(pad)
		out.WriteString(line)
		out.WriteString("\n")
	}
	out.WriteString(errorBlock)
	var content string
	if v.formReady {
		content = v.formViewport.View()
	} else {
		content = body.String()
	}
	for _, line := range strings.Split(content, "\n") {
		out.WriteString(pad)
		out.WriteString(line)
		out.WriteString("\n")
	}
	for _, line := range strings.Split(helpText, "\n") {
		out.WriteString(pad)
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}
func renderWrappedSelectableOptions(options []string, selectedIndex int, focused bool, maxWidth int) (string, int) {
	if len(options) == 0 {
		return "", 0
	}
	if maxWidth <= 0 {
		maxWidth = 80
	}
	lines := make([]string, 0, len(options))
	var currentLine string
	currentWidth := 0
	for i, option := range options {
		part := renderSelectableOption(option, i == selectedIndex, focused)
		partWidth := lipgloss.Width(part)
		separatorWidth := 0
		if currentLine != "" {
			separatorWidth = 2
		}
		if currentLine != "" && currentWidth+separatorWidth+partWidth > maxWidth {
			lines = append(lines, currentLine)
			currentLine = part
			currentWidth = partWidth
			continue
		}
		if currentLine != "" {
			currentLine += "  "
			currentWidth += separatorWidth
		}
		currentLine += part
		currentWidth += partWidth
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return strings.Join(lines, "\n  "), len(lines)
}
func (v *ConnectionView) renderWrappedSelectableOptions(options []string, selectedIndex int, focused bool, maxWidth int) (string, int) {
	key := v.selectorCacheKey(options, selectedIndex, focused, maxWidth)
	if cached, ok := v.selectorCache[key]; ok {
		return cached.rendered, cached.lineCount
	}
	rendered, lineCount := renderWrappedSelectableOptions(options, selectedIndex, focused, maxWidth)
	v.selectorCache[key] = wrappedSelectableOptionsCacheEntry{
		rendered:  rendered,
		lineCount: lineCount,
	}
	return rendered, lineCount
}
func (v *ConnectionView) selectorCacheKey(options []string, selectedIndex int, focused bool, maxWidth int) string {
	focusedKey := "0"
	if focused {
		focusedKey = "1"
	}
	themeName := ""
	if theme := styles.GetTheme(); theme != nil {
		themeName = theme.Name
	}
	return strings.Join([]string{
		themeName,
		strconv.Itoa(maxWidth),
		strconv.Itoa(selectedIndex),
		focusedKey,
		strings.Join(options, "\x1f"),
	}, "\x00")
}
func renderSelectableOption(option string, selected bool, focused bool) string {
	if selected {
		if focused {
			return styles.ActiveListItemStyle.Render(option)
		}
		return styles.RenderKey(option)
	}
	return styles.RenderMuted(option)
}
func (v *ConnectionView) selectorWrapWidth() int {
	width := v.width
	if width <= 0 {
		width = v.parent.width
	}
	if width <= 0 {
		width = 80
	}
	wrapWidth := width - 8
	if wrapWidth < 20 {
		return 20
	}
	return wrapWidth
}
func (v *ConnectionView) dbTypeSectionHeight() int {
	_, lineCount := v.renderWrappedSelectableOptions(v.dbTypes, v.dbTypeIndex, v.focusIndex == focusDBType, v.selectorWrapWidth())
	if lineCount == 0 {
		lineCount = 1
	}
	return lineCount + 2
}
func (v *ConnectionView) sslModeSectionHeight() int {
	sslModes := v.sslModes()
	if len(sslModes) == 0 {
		return 0
	}
	_, lineCount := v.renderWrappedSelectableOptions(sslModes, v.sslModeIndex, v.focusIndex == focusSSLMode, v.selectorWrapWidth())
	if lineCount == 0 {
		lineCount = 1
	}
	return lineCount + 2
}
func (v *ConnectionView) refreshList() {
	selectedName := v.selectedConnectionName()
	v.dockerItems = nil
	v.cloudItems = nil
	v.setListItems(selectedName)
	for k := range v.pingResults {
		delete(v.pingResults, k)
	}
}
func (v *ConnectionView) pingNextConnection() tea.Cmd {
	if len(v.pingQueue) == 0 {
		return nil
	}
	conn := v.pingQueue[0]
	v.pingQueue = v.pingQueue[1:]
	mgr := v.parent.dbManager
	return func() tea.Msg {
		online := mgr.Ping(&conn)
		return connectionPingMsg{name: conn.Name, online: online}
	}
}
func (v *ConnectionView) loadDockerConnections() tea.Cmd {
	return func() tea.Msg {
		return dockerConnectionsLoadedMsg{
			items: detectedConnectionItems(docker.DetectContainers()),
		}
	}
}
func (v *ConnectionView) loadCloudConnections() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		connections, err := cloudruntime.ListConnections(ctx, "")
		if err != nil && len(connections) == 0 {
			return cloudConnectionsLoadedMsg{}
		}
		return cloudConnectionsLoadedMsg{
			items: discoveredConnectionItems(connections),
		}
	}
}
func (v *ConnectionView) selectedConnectionName() string {
	item, ok := v.list.SelectedItem().(connectionItem)
	if !ok {
		return ""
	}
	return item.conn.Name
}
func (v *ConnectionView) setListItems(selectedName string) {
	items := baseConnectionItems(v.parent.dbManager.ListConnectionsWithSource())
	items = append(items, v.dockerItems...)
	items = append(items, v.cloudItems...)
	listItems := make([]list.Item, len(items))
	selectedIndex := 0
	for i, item := range items {
		listItems[i] = item
		if selectedName != "" && item.conn.Name == selectedName {
			selectedIndex = i
		}
	}
	v.list.SetItems(listItems)
	if len(listItems) > 0 {
		v.list.Select(selectedIndex)
	}
}
func (v *ConnectionView) formHasUserInput() bool {
	for _, input := range v.inputs {
		if input.Value() != "" {
			return true
		}
	}
	return false
}
func baseConnectionItems(infos []dbmgr.ConnectionSourceInfo) []connectionItem {
	items := make([]connectionItem, 0, len(infos))
	for _, info := range infos {
		items = append(items, connectionItem{conn: info.Connection, source: info.Source})
	}
	return items
}
func detectedConnectionItems(containers []docker.DetectedContainer) []connectionItem {
	items := make([]connectionItem, 0, len(containers))
	for _, c := range containers {
		items = append(items, connectionItem{
			conn: config.Connection{
				Name: c.Name,
				Type: c.Type,
				Host: "localhost",
				Port: c.Port,
			},
			source: ConnectionSourceDocker,
		})
	}
	return items
}
func discoveredConnectionItems(connections []cloudruntime.ConnectionSummary) []connectionItem {
	items := make([]connectionItem, 0, len(connections))
	for _, summary := range connections {
		conn, err := cloudruntime.BuildPrefillConnection(summary)
		if err != nil {
			continue
		}
		desc := summary.SourceType
		if conn.Host != "" {
			desc = fmt.Sprintf("%s@%s", summary.SourceType, conn.Host)
		}
		desc += " (cloud"
		if summary.ProviderType != "" {
			desc += "/" + summary.ProviderType
		}
		if summary.Region != "" {
			desc += " • " + summary.Region
		}
		desc += ")"
		items = append(items, connectionItem{
			conn:        conn,
			source:      ConnectionSourceCloud,
			description: desc,
		})
	}
	return items
}
func (v *ConnectionView) getFocusOrder() []int {
	order := []int{focusDBType}
	order = append(order, v.visibleFields...)
	if isNetworkDatabase(v.dbTypes[v.dbTypeIndex]) {
		order = append(order, focusSSHToggle)
		if v.sshEnabled {
			order = append(order, fieldSSHHost, fieldSSHUser, fieldSSHKeyFile, fieldSSHPassword)
		}
	}
	if len(v.sslModes()) > 0 {
		order = append(order, focusSSLMode)
		if v.sslFieldsVisible() {
			order = append(order, fieldSSLCAFile, fieldSSLCertFile, fieldSSLKeyFile, fieldSSLServer)
		}
	}
	order = append(order, focusConnect)
	return order
}
func (v *ConnectionView) nextInput() {
	if v.focusIndex < len(v.inputs) {
		v.inputs[v.focusIndex].Blur()
	}
	order := v.getFocusOrder()
	currentPos := -1
	for i, idx := range order {
		if idx == v.focusIndex {
			currentPos = i
			break
		}
	}
	nextPos := (currentPos + 1) % len(order)
	v.focusIndex = order[nextPos]
	if v.focusIndex < len(v.inputs) {
		v.inputs[v.focusIndex].Focus()
	}
	v.scrollToFocused()
}
func (v *ConnectionView) prevInput() {
	if v.focusIndex < len(v.inputs) {
		v.inputs[v.focusIndex].Blur()
	}
	order := v.getFocusOrder()
	currentPos := -1
	for i, idx := range order {
		if idx == v.focusIndex {
			currentPos = i
			break
		}
	}
	prevPos := (currentPos - 1 + len(order)) % len(order)
	v.focusIndex = order[prevPos]
	if v.focusIndex < len(v.inputs) {
		v.inputs[v.focusIndex].Focus()
	}
	v.scrollToFocused()
}
func (v *ConnectionView) scrollToFocused() {
	if !v.formReady {
		return
	}
	line := 0
	if v.focusIndex == focusDBType {
		v.formViewport.GotoTop()
		return
	}
	line += v.dbTypeSectionHeight()
	for _, idx := range v.visibleFields {
		if idx == v.focusIndex {
			break
		}
		line += 3
	}
	if isNetworkDatabase(v.dbTypes[v.dbTypeIndex]) {
		if v.focusIndex == focusSSHToggle {
		} else {
			line += 3
			if v.sshEnabled {
				for _, idx := range []int{fieldSSHHost, fieldSSHUser, fieldSSHKeyFile, fieldSSHPassword} {
					if idx == v.focusIndex {
						break
					}
					line += 3
				}
			}
		}
	}
	if len(v.sslModes()) > 0 {
		if v.focusIndex == focusSSLMode {
		} else {
			line += v.sslModeSectionHeight()
			if v.sslFieldsVisible() {
				for _, idx := range []int{fieldSSLCAFile, fieldSSLCertFile, fieldSSLKeyFile, fieldSSLServer} {
					if idx == v.focusIndex {
						break
					}
					line += 3
				}
			}
		}
	}
	if v.focusIndex == focusConnect {
		v.formViewport.GotoBottom()
		return
	}
	vpHeight := v.formViewport.Height
	offset := v.formViewport.YOffset
	if line < offset+1 {
		newOffset := line - 1
		if newOffset < 0 {
			newOffset = 0
		}
		v.formViewport.SetYOffset(newOffset)
	} else if line+2 >= offset+vpHeight {
		newOffset := line - vpHeight + 3
		if newOffset < 0 {
			newOffset = 0
		}
		v.formViewport.SetYOffset(newOffset)
	}
}
func (v *ConnectionView) resetForm() {
	for i := range v.inputs {
		v.inputs[i].SetValue("")
		v.inputs[i].Blur()
	}
	v.advanced = nil
	v.focusIndex = focusDBType
	v.dbTypeIndex = 0
	v.sshEnabled = false
	v.sslModeIndex = 0
	v.connError = nil
	v.onDbTypeChanged()
}
func (v *ConnectionView) prefillFromConnection(conn config.Connection) {
	for i, t := range v.dbTypes {
		if strings.EqualFold(t, conn.Type) {
			v.dbTypeIndex = i
			break
		}
	}
	v.onDbTypeChanged()
	if conn.Name != "" {
		v.inputs[fieldName].SetValue(conn.Name)
	}
	if conn.Host != "" {
		v.inputs[fieldHost].SetValue(conn.Host)
	}
	if conn.Port > 0 {
		v.inputs[fieldPort].SetValue(strconv.Itoa(conn.Port))
	}
	if conn.Username != "" {
		v.inputs[fieldUsername].SetValue(conn.Username)
	}
	if conn.Database != "" {
		v.inputs[fieldDatabase].SetValue(conn.Database)
	}
	if conn.Schema != "" {
		v.inputs[fieldSchema].SetValue(conn.Schema)
	}
	v.advanced = cloneAdvanced(conn.Advanced)
	if conn.SSHHost != "" {
		v.sshEnabled = true
		v.inputs[fieldSSHHost].SetValue(conn.SSHHost)
		if conn.SSHUser != "" {
			v.inputs[fieldSSHUser].SetValue(conn.SSHUser)
		}
		if conn.SSHKeyFile != "" {
			v.inputs[fieldSSHKeyFile].SetValue(conn.SSHKeyFile)
		}
	}
	sslSettings := connectionopts.SSLSettingsFromAdvanced(conn.Advanced)
	if sslSettings.Mode != "" {
		for i, mode := range v.sslModes() {
			if strings.EqualFold(mode, sslSettings.Mode) {
				v.sslModeIndex = i
				break
			}
		}
	}
	if sslSettings.ServerName != "" {
		v.inputs[fieldSSLServer].SetValue(sslSettings.ServerName)
	}
	v.focusIndex = focusConnect
	for _, idx := range v.getFocusOrder() {
		if idx == focusDBType {
			continue
		}
		if idx >= 0 && idx < len(v.inputs) && strings.TrimSpace(v.inputs[idx].Value()) == "" {
			v.focusIndex = idx
			break
		}
	}
	if v.focusIndex >= 0 && v.focusIndex < len(v.inputs) {
		v.inputs[v.focusIndex].Focus()
	}
	v.scrollToFocused()
}
func cloneAdvanced(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
func (v *ConnectionView) updatePortPlaceholder() {
	defaultPort := v.getDefaultPort(v.dbTypes[v.dbTypeIndex])
	if defaultPort <= 0 {
		v.inputs[fieldPort].Placeholder = ""
		return
	}
	v.inputs[fieldPort].Placeholder = strconv.Itoa(defaultPort)
}
func (v *ConnectionView) getDefaultPort(dbType string) int {
	port, ok := sourcetypes.DefaultPort(dbType)
	if !ok {
		return 0
	}
	return port
}
func isNetworkDatabase(dbType string) bool {
	return sourcetypes.IsNetworkTransport(dbType)
}
func isFileBasedFormDatabase(dbType string) bool {
	return sourcetypes.IsFileTransport(dbType)
}
func (v *ConnectionView) sslModes() []string {
	modeInfos := sourcetypes.SSLModes(v.dbTypes[v.dbTypeIndex])
	modes := make([]string, 0, len(modeInfos))
	for _, mode := range modeInfos {
		modes = append(modes, string(mode.Value))
	}
	return modes
}
func (v *ConnectionView) currentSSLMode() string {
	modes := v.sslModes()
	if len(modes) == 0 {
		return ""
	}
	if v.sslModeIndex < 0 || v.sslModeIndex >= len(modes) {
		return modes[0]
	}
	return modes[v.sslModeIndex]
}
func (v *ConnectionView) sslFieldsVisible() bool {
	mode := v.currentSSLMode()
	return mode != "" && mode != "disabled"
}
func getVisibleFields(dbType string) []int {
	spec, ok := sourcetypes.Find(dbType)
	if !ok {
		return []int{fieldName, fieldHost, fieldPort, fieldUsername, fieldPassword, fieldDatabase}
	}
	fields := []int{fieldName}
	if _, ok := spec.ConnectionFieldByKey("Hostname"); ok {
		fields = append(fields, fieldHost, fieldPort)
	}
	if _, ok := spec.ConnectionFieldByKey("Username"); ok {
		fields = append(fields, fieldUsername)
	}
	if _, ok := spec.ConnectionFieldByKey("Password"); ok {
		fields = append(fields, fieldPassword)
	}
	if _, ok := spec.ConnectionFieldByKey("Database"); ok {
		fields = append(fields, fieldDatabase)
	}
	if _, ok := spec.ConnectionFieldByKey("Search Path"); ok {
		fields = append(fields, fieldSchema)
	}
	return fields
}
func passwordRequired(dbType string) bool {
	return sourcetypes.ConnectionFieldRequired(dbType, "Password")
}
func (v *ConnectionView) isFieldVisible(index int) bool {
	for _, vi := range v.visibleFields {
		if vi == index {
			return true
		}
	}
	return false
}
func (v *ConnectionView) isFocusableVisible(index int) bool {
	switch index {
	case focusDBType, focusConnect:
		return true
	case focusSSHToggle:
		return isNetworkDatabase(v.dbTypes[v.dbTypeIndex])
	case focusSSLMode:
		return len(v.sslModes()) > 0
	case fieldSSHHost, fieldSSHUser, fieldSSHKeyFile, fieldSSHPassword:
		return isNetworkDatabase(v.dbTypes[v.dbTypeIndex]) && v.sshEnabled
	case fieldSSLCAFile, fieldSSLCertFile, fieldSSLKeyFile, fieldSSLServer:
		return len(v.sslModes()) > 0 && v.sslFieldsVisible()
	default:
		return v.isFieldVisible(index)
	}
}
func (v *ConnectionView) onDbTypeChanged() {
	v.updatePortPlaceholder()
	v.visibleFields = getVisibleFields(v.dbTypes[v.dbTypeIndex])
	v.advanced = nil
	sslModes := v.sslModes()
	v.sslModeIndex = 0
	for i, mode := range sslModes {
		if mode == "disabled" {
			v.sslModeIndex = i
			break
		}
	}
	if isFileBasedFormDatabase(v.dbTypes[v.dbTypeIndex]) {
		v.inputs[fieldDatabase].Placeholder = "/path/to/database.db"
	} else {
		v.inputs[fieldDatabase].Placeholder = "mydb"
	}
	if !isNetworkDatabase(v.dbTypes[v.dbTypeIndex]) {
		v.sshEnabled = false
	}
	if len(sslModes) == 0 {
		for _, idx := range []int{fieldSSLCAFile, fieldSSLCertFile, fieldSSLKeyFile, fieldSSLServer} {
			v.inputs[idx].SetValue("")
		}
	}
	if !v.isFocusableVisible(v.focusIndex) {
		v.nextInput()
	}
}
func (v *ConnectionView) connect() tea.Cmd {
	name := v.inputs[fieldName].Value()
	dbType := v.dbTypes[v.dbTypeIndex]
	baseAdvanced := cloneAdvanced(v.advanced)
	host := ""
	if v.isFieldVisible(fieldHost) {
		host = v.inputs[fieldHost].Value()
	}
	if host == "" {
		if database := v.inputs[fieldDatabase].Value(); isFileBasedFormDatabase(dbType) && database != "" {
			host = database
		} else if sourcetypes.ConnectionFieldRequired(dbType, "Hostname") {
			host = "localhost"
		}
	}
	var port int
	if v.isFieldVisible(fieldPort) {
		portStr := v.inputs[fieldPort].Value()
		if portStr == "" {
			port = v.getDefaultPort(dbType)
		} else {
			portNum, err := strconv.Atoi(portStr)
			if err != nil || portNum < 1 || portNum > 65535 {
				return func() tea.Msg {
					return connectionResultMsg{err: fmt.Errorf("invalid port number: must be between 1 and 65535")}
				}
			}
			port = portNum
		}
	} else if defaultPort, ok := sourcetypes.DefaultPort(dbType); ok {
		port = defaultPort
	}
	username := ""
	if v.isFieldVisible(fieldUsername) {
		username = v.inputs[fieldUsername].Value()
	}
	password := ""
	if v.isFieldVisible(fieldPassword) {
		password = v.inputs[fieldPassword].Value()
	}
	database := ""
	if v.isFieldVisible(fieldDatabase) {
		database = v.inputs[fieldDatabase].Value()
	}
	schema := ""
	if v.isFieldVisible(fieldSchema) {
		schema = v.inputs[fieldSchema].Value()
	}
	conn := config.Connection{
		Name:     name,
		Type:     dbType,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
		Schema:   schema,
	}
	sslMode := ""
	if v.currentSSLMode() != "disabled" || strings.TrimSpace(v.inputs[fieldSSLCAFile].Value()) != "" ||
		strings.TrimSpace(v.inputs[fieldSSLCertFile].Value()) != "" || strings.TrimSpace(v.inputs[fieldSSLKeyFile].Value()) != "" ||
		strings.TrimSpace(v.inputs[fieldSSLServer].Value()) != "" {
		sslMode = v.currentSSLMode()
	}
	advanced, err := connectionopts.ApplySSLSettings(dbType, baseAdvanced, connectionopts.SSLSettings{
		Mode:           sslMode,
		CAFile:         v.inputs[fieldSSLCAFile].Value(),
		ClientCertFile: v.inputs[fieldSSLCertFile].Value(),
		ClientKeyFile:  v.inputs[fieldSSLKeyFile].Value(),
		ServerName:     v.inputs[fieldSSLServer].Value(),
	})
	if err != nil {
		return func() tea.Msg {
			return connectionResultMsg{err: err}
		}
	}
	conn.Advanced = advanced
	if v.sshEnabled && isNetworkDatabase(dbType) {
		conn.SSHHost = v.inputs[fieldSSHHost].Value()
		conn.SSHUser = v.inputs[fieldSSHUser].Value()
		conn.SSHKeyFile = v.inputs[fieldSSHKeyFile].Value()
		conn.SSHPassword = v.inputs[fieldSSHPassword].Value()
		if portStr := v.inputs[fieldPort].Value(); portStr != "" {
		}
	}
	dbManager := v.parent.dbManager
	cfg := v.parent.config
	return func() tea.Msg {
		if err := dbManager.Connect(&conn); err != nil {
			return connectionResultMsg{err: err}
		}
		sslSummary, err := dbManager.GetSSLStatusSummary()
		if err != nil {
			sslSummary = ""
		}
		if name != "" {
			cfg.AddConnection(conn)
			cfg.Save()
		}
		return connectionResultMsg{
			err:           nil,
			statusMessage: buildConnectionStatusMessage(&conn, sslSummary),
		}
	}
}