package types
import (
	"encoding/json"
	"github.com/clidey/whodb/core/src/migrate"
)
type DatabaseCredentials struct {
	Alias    string            `json:"alias"`
	Hostname string            `json:"host"`
	Username string            `json:"user"`
	Password string            `json:"password"`
	Database string            `json:"database"`
	Port     string            `json:"port"`
	Advanced map[string]string `json:"advanced"`
	Extra    map[string]any
	IsProfile bool
	Type      string
	CustomId  string
	Source    string
}
func (d *DatabaseCredentials) UnmarshalJSON(data []byte) error {
	type Alias DatabaseCredentials
	aux := &struct {
		*Alias
		Config map[string]string `json:"config"`
	}{Alias: (*Alias)(d)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if d.Advanced == nil && aux.Config != nil {
		d.Advanced = aux.Config
		migrate.DeprecatedConfigKey = true
	}
	return nil
}