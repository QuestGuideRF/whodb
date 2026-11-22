package settings
import "github.com/clidey/whodb/core/src/analytics"
type Settings struct {
	MetricsEnabled bool `json:"metricsEnabled"`
}
type ISettingsField interface {
	Apply(*Settings) bool
}
type MetricsEnabledField bool
var currentSettings = Settings{MetricsEnabled: false}
func Get() Settings {
	return currentSettings
}
func (m MetricsEnabledField) Apply(s *Settings) bool {
	enabled := bool(m)
	if s.MetricsEnabled == enabled {
		analytics.SetEnabled(enabled)
		return false
	}
	s.MetricsEnabled = enabled
	analytics.SetEnabled(enabled)
	return true
}
func UpdateSettings(fields ...ISettingsField) bool {
	changed := false
	for _, field := range fields {
		if field.Apply(&currentSettings) {
			changed = true
		}
	}
	return changed
}