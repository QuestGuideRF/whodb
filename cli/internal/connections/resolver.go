package connections
import (
	"fmt"
	"strconv"
	"strings"
	"github.com/clidey/whodb/cli/internal/config"
	"github.com/clidey/whodb/cli/internal/sourcetypes"
	"github.com/clidey/whodb/core/src/envconfig"
	"github.com/clidey/whodb/core/src/types"
)
const ConnectionSourceSaved = "saved"
const ConnectionSourceEnv = "env"
type ConnectionSourceInfo struct {
	Connection config.Connection
	Source     string
}
type Resolver struct {
	infos []ConnectionSourceInfo
}
func NewResolver(includeSecrets bool) (*Resolver, error) {
	var (
		cfg *config.Config
		err error
	)
	if includeSecrets {
		cfg, err = config.LoadConfig()
	} else {
		cfg, err = config.LoadConfigWithoutSecrets()
	}
	if err != nil {
		return nil, err
	}
	return NewResolverWithConfig(cfg), nil
}
func NewResolverWithConfig(cfg *config.Config) *Resolver {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &Resolver{
		infos: listConnectionsWithSource(cfg),
	}
}
func (r *Resolver) Count() int {
	if r == nil {
		return 0
	}
	return len(r.infos)
}
func (r *Resolver) ListWithSource() []ConnectionSourceInfo {
	if r == nil {
		return nil
	}
	infos := make([]ConnectionSourceInfo, len(r.infos))
	copy(infos, r.infos)
	return infos
}
func (r *Resolver) List() []config.Connection {
	if r == nil {
		return nil
	}
	conns := make([]config.Connection, len(r.infos))
	for i, info := range r.infos {
		conns[i] = info.Connection
	}
	return conns
}
func (r *Resolver) Resolve(name string) (*config.Connection, string, error) {
	if strings.TrimSpace(name) == "" {
		return nil, "", fmt.Errorf("connection name is required")
	}
	if r == nil {
		return nil, "", fmt.Errorf("connection %q not found", name)
	}
	for _, info := range r.infos {
		if info.Connection.Name == name {
			conn := info.Connection
			return &conn, info.Source, nil
		}
	}
	return nil, "", fmt.Errorf("connection %q not found", name)
}
func EnvConnections() []config.Connection {
	typeCounts := make(map[string]int)
	connections := make([]config.Connection, 0)
	for _, dbType := range sourcetypes.IDs() {
		profiles := envconfig.GetDefaultDatabaseCredentials(dbType)
		for _, profile := range profiles {
			typeCounts[dbType]++
			connections = append(connections, envProfileToConnection(profile, dbType, typeCounts[dbType]))
		}
	}
	return connections
}
func listConnectionsWithSource(cfg *config.Config) []ConnectionSourceInfo {
	saved := cfg.Connections
	envConnections := EnvConnections()
	infos := make([]ConnectionSourceInfo, 0, len(saved)+len(envConnections))
	usedNames := make(map[string]bool, len(saved)+len(envConnections))
	for _, conn := range saved {
		infos = append(infos, ConnectionSourceInfo{
			Connection: conn,
			Source:     ConnectionSourceSaved,
		})
		usedNames[conn.Name] = true
	}
	for _, conn := range envConnections {
		if conn.Name == "" || usedNames[conn.Name] {
			continue
		}
		infos = append(infos, ConnectionSourceInfo{
			Connection: conn,
			Source:     ConnectionSourceEnv,
		})
		usedNames[conn.Name] = true
	}
	return infos
}
func envProfileToConnection(profile types.DatabaseCredentials, dbType string, index int) config.Connection {
	name := envProfileName(profile, dbType, index)
	var advanced map[string]string
	if profile.Port != "" || len(profile.Advanced) > 0 {
		advanced = make(map[string]string, len(profile.Advanced)+1)
		for key, value := range profile.Advanced {
			advanced[key] = value
		}
		if profile.Port != "" {
			advanced["Port"] = profile.Port
		}
	}
	port := 0
	if profile.Port != "" {
		if parsedPort, err := strconv.Atoi(profile.Port); err == nil {
			port = parsedPort
		}
	}
	return config.Connection{
		Name:      name,
		Type:      dbType,
		Host:      profile.Hostname,
		Port:      port,
		Username:  profile.Username,
		Password:  profile.Password,
		Database:  profile.Database,
		Advanced:  advanced,
		IsProfile: true,
	}
}
func envProfileName(profile types.DatabaseCredentials, dbType string, index int) string {
	if profile.CustomId != "" {
		return profile.CustomId
	}
	if profile.Alias != "" {
		return profile.Alias
	}
	return fmt.Sprintf("%s-%d", strings.ToLower(dbType), index)
}