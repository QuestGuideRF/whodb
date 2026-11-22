package env
type GCPProviderEnvConfig struct {
	Name string `json:"name"`
	ProjectID string `json:"projectId"`
	Region string `json:"region"`
	ServiceAccountKeyPath string `json:"serviceAccountKeyPath,omitempty"`
	DiscoverCloudSQL *bool `json:"discoverCloudSQL,omitempty"`
	DiscoverAlloyDB *bool `json:"discoverAlloyDB,omitempty"`
	DiscoverMemorystore *bool `json:"discoverMemorystore,omitempty"`
}