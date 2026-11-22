package src
import (
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/source"
	"github.com/clidey/whodb/core/src/types"
)
func GetSourceProfiles() []source.Profile {
	loginProfiles := GetLoginProfiles()
	profiles := make([]source.Profile, 0, len(loginProfiles))
	for i, profile := range loginProfiles {
		id := GetLoginProfileId(i, profile)
		displayName := id
		if profile.Alias != "" {
			displayName = profile.Alias
		}
		if profile.CustomId != "" {
			id = profile.CustomId
		}
		sslConfigured := false
		if mode, ok := profile.Advanced[ssl.KeySSLMode]; ok && mode != "" && mode != string(ssl.SSLModeDisabled) {
			sslConfigured = true
		}
		profiles = append(profiles, source.Profile{
			ID:                   id,
			DisplayName:          displayName,
			SourceType:           profile.Type,
			Values:               displayValuesForProfile(profile),
			IsEnvironmentDefined: profile.Source == "environment",
			Source:               profile.Source,
			SSLConfigured:        sslConfigured,
		})
	}
	return profiles
}
func FindSourceProfile(id string) (*source.Profile, *source.Credentials, bool) {
	loginProfiles := GetLoginProfiles()
	for i, profile := range loginProfiles {
		profileID := GetLoginProfileId(i, profile)
		if profile.CustomId != "" {
			profileID = profile.CustomId
		}
		if profileID != id {
			continue
		}
		displayName := profileID
		if profile.Alias != "" {
			displayName = profile.Alias
		}
		sourceProfile := &source.Profile{
			ID:                   profileID,
			DisplayName:          displayName,
			SourceType:           profile.Type,
			Values:               displayValuesForProfile(profile),
			IsEnvironmentDefined: profile.Source == "environment",
			Source:               profile.Source,
		}
		credentials := GetSourceCredentials(profile)
		credentials.ID = &sourceProfile.ID
		return sourceProfile, credentials, true
	}
	return nil, nil, false
}
func GetSourceCredentials(profile types.DatabaseCredentials) *source.Credentials {
	values := make(map[string]string, len(profile.Advanced)+4)
	if profile.Hostname != "" {
		values["Hostname"] = profile.Hostname
	}
	if profile.Username != "" {
		values["Username"] = profile.Username
	}
	if profile.Password != "" {
		values["Password"] = profile.Password
	}
	if profile.Database != "" {
		values["Database"] = profile.Database
	}
	if profile.Port != "" {
		values["Port"] = profile.Port
	}
	for key, value := range profile.Advanced {
		values[key] = value
	}
	return &source.Credentials{
		SourceType: profile.Type,
		Values:     values,
		IsProfile:  profile.IsProfile,
	}
}
func displayValuesForProfile(profile types.DatabaseCredentials) map[string]string {
	values := map[string]string{}
	if profile.Hostname != "" {
		values["Hostname"] = profile.Hostname
	}
	if profile.Database != "" {
		values["Database"] = profile.Database
	}
	if profile.Port != "" {
		values["Port"] = profile.Port
	}
	for key, value := range profile.Advanced {
		if key == ssl.KeySSLCACertContent || key == ssl.KeySSLClientCertContent || key == ssl.KeySSLClientKeyContent {
			continue
		}
		values[key] = value
	}
	return values
}