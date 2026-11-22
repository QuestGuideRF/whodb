package env
type AWSProviderEnvConfig struct {
	Name string `json:"name"`
	Region string `json:"region"`
	ProfileName string `json:"profileName,omitempty"`
	DiscoverRDS *bool `json:"discoverRDS,omitempty"`
	DiscoverElastiCache *bool `json:"discoverElastiCache,omitempty"`
	DiscoverDocumentDB *bool `json:"discoverDocumentDB,omitempty"`
}