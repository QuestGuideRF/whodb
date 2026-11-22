package gcp
import (
	"context"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"google.golang.org/api/option"
)
type GCPConfig struct {
	ProjectID string
	Region    string
	Options   []option.ClientOption
}
func LoadGCPConfig(ctx context.Context, creds *engine.Credentials) (*GCPConfig, error) {
	gcpCreds, err := ParseFromWhoDB(creds)
	if err != nil {
		return nil, err
	}
	return loadConfigFromGCPCredentials(ctx, gcpCreds)
}
func loadConfigFromGCPCredentials(ctx context.Context, gcpCreds *GCPCredentialConfig) (*GCPConfig, error) {
	var opts []option.ClientOption
	if gcpCreds.IsServiceAccountKeyAuth() && gcpCreds.ServiceAccountKeyPath != "" {
		opts = append(opts, option.WithCredentialsFile(gcpCreds.ServiceAccountKeyPath))
	}
	log.WithFields(map[string]any{
		"projectID":  gcpCreds.ProjectID,
		"region":     gcpCreds.Region,
		"authMethod": gcpCreds.AuthMethod,
	}).Debug("GCP configuration loaded successfully")
	return &GCPConfig{
		ProjectID: gcpCreds.ProjectID,
		Region:    gcpCreds.Region,
		Options:   opts,
	}, nil
}