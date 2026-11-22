package config
import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
	"github.com/clidey/whodb/cli/pkg/identity"
	"github.com/clidey/whodb/core/graph/model"
	"github.com/clidey/whodb/core/src/common/config"
	"github.com/clidey/whodb/core/src/common/datadir"
	"github.com/clidey/whodb/core/src/env"
	"github.com/zalando/go-keyring"
)
var (
	globalUseKeyring bool
	keyringAvailable *bool
)
type Connection struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Username    string            `json:"username"`
	Password    string            `json:"password,omitempty"`
	Database    string            `json:"database"`
	Schema      string            `json:"schema,omitempty"`
	Advanced    map[string]string `json:"advanced,omitempty"`
	IsProfile   bool              `json:"is_profile,omitempty"`
	SSHHost     string            `json:"ssh_host,omitempty"`
	SSHPort     int               `json:"ssh_port,omitempty"`
	SSHUser     string            `json:"ssh_user,omitempty"`
	SSHKeyFile  string            `json:"ssh_key_file,omitempty"`
	SSHPassword string            `json:"ssh_password,omitempty"`
}
type HistoryConfig struct {
	MaxEntries int  `json:"max_entries"`
	Persist    bool `json:"persist"`
}
type DisplayConfig struct {
	Theme    string `json:"theme"`
	PageSize int    `json:"page_size"`
}
type AIConfig struct {
	ConsentGiven bool   `json:"consent_given"`
	LastProvider string `json:"last_provider,omitempty"`
	LastModel    string `json:"last_model,omitempty"`
}
type QueryConfig struct {
	TimeoutSeconds          int `json:"timeout_seconds"`
	PreferredTimeoutSeconds int `json:"preferred_timeout_seconds,omitempty"`
}
type Profile struct {
	Name           string `json:"name"`
	Connection     string `json:"connection"`
	Theme          string `json:"theme,omitempty"`
	PageSize       int    `json:"page_size,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}
type SavedQuery struct {
	Name  string `json:"name"`
	Query string `json:"query"`
}
type WorkspaceEditorBufferState struct {
	Name  string `json:"name"`
	Query string `json:"query"`
}
type WorkspaceEditorState struct {
	Buffers   []WorkspaceEditorBufferState `json:"buffers,omitempty"`
	ActiveTab int                          `json:"active_tab,omitempty"`
}
type WorkspaceBrowserState struct {
	Schema string `json:"schema,omitempty"`
	Table  string `json:"table,omitempty"`
	Filter string `json:"filter,omitempty"`
}
type WorkspaceResultsState struct {
	Schema         string                `json:"schema,omitempty"`
	Table          string                `json:"table,omitempty"`
	CurrentPage    int                   `json:"current_page,omitempty"`
	PageSize       int                   `json:"page_size,omitempty"`
	ColumnOffset   int                   `json:"column_offset,omitempty"`
	VisibleColumns []string              `json:"visible_columns,omitempty"`
	Where          *model.WhereCondition `json:"where,omitempty"`
}
type WorkspaceDiffState struct {
	FromConnection string `json:"from_connection,omitempty"`
	ToConnection   string `json:"to_connection,omitempty"`
	FromSchema     string `json:"from_schema,omitempty"`
	ToSchema       string `json:"to_schema,omitempty"`
}
type WorkspaceState struct {
	ConnectionName string                `json:"connection_name,omitempty"`
	ProfileName    string                `json:"profile_name,omitempty"`
	View           string                `json:"view,omitempty"`
	Layout         string                `json:"layout,omitempty"`
	FocusedPane    int                   `json:"focused_pane,omitempty"`
	Browser        WorkspaceBrowserState `json:"browser,omitempty"`
	Editor         WorkspaceEditorState  `json:"editor,omitempty"`
	Results        WorkspaceResultsState `json:"results,omitempty"`
	Diff           WorkspaceDiffState    `json:"diff,omitempty"`
	SavedAt        string                `json:"saved_at,omitempty"`
}
type CLISection struct {
	Connections  []Connection    `json:"connections"`
	History      HistoryConfig   `json:"history"`
	Display      DisplayConfig   `json:"display"`
	AI           AIConfig        `json:"ai"`
	Query        QueryConfig     `json:"query"`
	SavedQueries []SavedQuery    `json:"saved_queries,omitempty"`
	Profiles     []Profile       `json:"profiles,omitempty"`
	ReadOnly     bool            `json:"read_only,omitempty"`
	Workspace    *WorkspaceState `json:"workspace,omitempty"`
}
type Config struct {
	CLISection
	useKeyring          bool
	keyringWarningShown bool
}
func DefaultConfig() *Config {
	return &Config{
		CLISection: CLISection{
			Connections: []Connection{},
			History: HistoryConfig{
				MaxEntries: 1000,
				Persist:    true,
			},
			Display: DisplayConfig{
				Theme:    "dark",
				PageSize: 50,
			},
			AI: AIConfig{
				ConsentGiven: false,
			},
			Query: QueryConfig{
				TimeoutSeconds: 30,
			},
		},
		useKeyring: false,
	}
}
var (
	configDir     string
	configDirOnce sync.Once
	configDirErr  error
)
func ResetPathsForTesting() {
	configDir = ""
	configDirErr = nil
	configDirOnce = sync.Once{}
	config.ResetConfigPath()
}
func getConfigOptions() datadir.Options {
	return datadir.Options{
		AppName:           "whodb",
		EnterpriseEdition: env.IsEnterpriseEdition,
		Development:       env.IsDevelopment,
	}
}
func GetConfigDir() (string, error) {
	configDirOnce.Do(func() {
		configDir, configDirErr = datadir.Get(getConfigOptions())
		if configDirErr != nil {
			return
		}
	})
	return configDir, configDirErr
}
func GetConfigPath() (string, error) {
	return config.GetPath(getConfigOptions())
}
func isKeyringAvailable() bool {
	if keyringAvailable != nil {
		return *keyringAvailable
	}
	_, err := keyring.Get(identity.Current().KeyringService, "whodb-cli-test-availability")
	available := err == nil || errors.Is(err, keyring.ErrNotFound)
	keyringAvailable = &available
	return available
}
func LoadConfig() (*Config, error) {
	return loadConfig(true, true)
}
func LoadConfigWithoutSecrets() (*Config, error) {
	return loadConfig(false, false)
}
func loadConfig(includeSecrets, showWarnings bool) (*Config, error) {
	if _, err := GetConfigDir(); err != nil {
		return nil, err
	}
	useKeyring := isKeyringAvailable()
	globalUseKeyring = useKeyring
	cfg := DefaultConfig()
	cfg.useKeyring = useKeyring
	opts := getConfigOptions()
	if err := config.ReadSection(config.SectionCLI, &cfg.CLISection, opts); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}
	if includeSecrets && cfg.useKeyring {
		for i := range cfg.Connections {
			if cfg.Connections[i].Name != "" {
				password, err := keyring.Get(identity.Current().KeyringService, "connection:"+cfg.Connections[i].Name)
				if err == nil {
					cfg.Connections[i].Password = password
				}
			}
		}
	}
	if showWarnings {
		cfg.showKeyringWarning()
	}
	return cfg, nil
}
func (c *Config) showKeyringWarning() {
	if !c.useKeyring && !c.keyringWarningShown {
		fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: OS keyring not available.\n")
		fmt.Fprintf(os.Stderr, "   Passwords will be stored in plaintext in config file.\n")
		fmt.Fprintf(os.Stderr, "   File permissions: 0600 (user read/write only)\n\n")
		c.keyringWarningShown = true
	}
}
func (c *Config) Save() error {
	globalUseKeyring = c.useKeyring
	if c.useKeyring {
		for _, conn := range c.Connections {
			if conn.Name != "" && conn.Password != "" {
				err := keyring.Set(identity.Current().KeyringService, "connection:"+conn.Name, conn.Password)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Could not save password to keyring for %s: %v\n", conn.Name, err)
					fmt.Fprintf(os.Stderr, "Password will be saved in config file.\n")
					globalUseKeyring = false
				}
			}
		}
	}
	section := c.CLISection
	if globalUseKeyring {
		section.Connections = make([]Connection, len(c.Connections))
		for i, conn := range c.Connections {
			section.Connections[i] = conn
			section.Connections[i].Password = ""
		}
	}
	opts := getConfigOptions()
	if err := config.WriteSection(config.SectionCLI, section, opts); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}
	return nil
}
func (c *Config) AddConnection(conn Connection) {
	for i, existing := range c.Connections {
		if existing.Name == conn.Name {
			c.Connections[i] = conn
			return
		}
	}
	c.Connections = append(c.Connections, conn)
}
func (c *Config) RemoveConnection(name string) bool {
	for i, conn := range c.Connections {
		if conn.Name == name {
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
			if c.useKeyring {
				_ = keyring.Delete(identity.Current().KeyringService, "connection:"+name)
			}
			return true
		}
	}
	return false
}
func (c *Config) GetConnection(name string) (*Connection, error) {
	for _, conn := range c.Connections {
		if conn.Name == name {
			return &conn, nil
		}
	}
	return nil, fmt.Errorf("connection '%s' not found", name)
}
func (c *Config) SetAIConsent(consent bool) {
	c.AI.ConsentGiven = consent
}
func (c *Config) GetAIConsent() bool {
	return c.AI.ConsentGiven
}
func (c *Config) GetQueryTimeout() time.Duration {
	if c.Query.TimeoutSeconds <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.Query.TimeoutSeconds) * time.Second
}
func (c *Config) GetLastAIProvider() string {
	return c.AI.LastProvider
}
func (c *Config) SetLastAIProvider(provider string) {
	c.AI.LastProvider = provider
}
func (c *Config) GetLastAIModel() string {
	return c.AI.LastModel
}
func (c *Config) SetLastAIModel(model string) {
	c.AI.LastModel = model
}
func (c *Config) GetPreferredTimeout() int {
	return c.Query.PreferredTimeoutSeconds
}
func (c *Config) SetPreferredTimeout(seconds int) {
	c.Query.PreferredTimeoutSeconds = seconds
}
func (c *Config) GetPageSize() int {
	if c.Display.PageSize <= 0 {
		return 50
	}
	return c.Display.PageSize
}
func (c *Config) SetPageSize(size int) {
	c.Display.PageSize = size
}
func (c *Config) GetThemeName() string {
	name := c.Display.Theme
	if name == "" || name == "dark" {
		return "default"
	}
	return name
}
func (c *Config) SetThemeName(name string) {
	c.Display.Theme = name
}
func (c *Config) AddSavedQuery(name, query string) {
	for i, sq := range c.SavedQueries {
		if sq.Name == name {
			c.SavedQueries[i].Query = query
			return
		}
	}
	c.SavedQueries = append(c.SavedQueries, SavedQuery{Name: name, Query: query})
}
func (c *Config) DeleteSavedQuery(name string) bool {
	for i, sq := range c.SavedQueries {
		if sq.Name == name {
			c.SavedQueries = append(c.SavedQueries[:i], c.SavedQueries[i+1:]...)
			return true
		}
	}
	return false
}
func (c *Config) GetSavedQueries() []SavedQuery {
	return c.SavedQueries
}
func (c *Config) GetReadOnly() bool {
	return c.ReadOnly
}
func (c *Config) SetReadOnly(readOnly bool) {
	c.ReadOnly = readOnly
}
func (c *Config) AddProfile(profile Profile) {
	for i, p := range c.Profiles {
		if p.Name == profile.Name {
			c.Profiles[i] = profile
			return
		}
	}
	c.Profiles = append(c.Profiles, profile)
}
func (c *Config) DeleteProfile(name string) bool {
	for i, p := range c.Profiles {
		if p.Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return true
		}
	}
	return false
}
func (c *Config) GetProfile(name string) *Profile {
	for _, p := range c.Profiles {
		if p.Name == name {
			return &p
		}
	}
	return nil
}
func (c *Config) GetProfiles() []Profile {
	return c.Profiles
}
func (c *Config) GetWorkspace() *WorkspaceState {
	return c.Workspace
}
func (c *Config) SetWorkspace(workspace *WorkspaceState) {
	c.Workspace = workspace
}
func (c *Config) ClearWorkspace() {
	c.Workspace = nil
}