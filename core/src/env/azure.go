package env
type AzureProviderEnvConfig struct {
	Name string `json:"name"`
	SubscriptionID string `json:"subscriptionId"`
	TenantID string `json:"tenantId,omitempty"`
	ClientID string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	AuthMethod string `json:"authMethod,omitempty"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
	DiscoverPostgreSQL *bool `json:"discoverPostgreSQL,omitempty"`
	DiscoverMySQL *bool `json:"discoverMySQL,omitempty"`
	DiscoverRedis *bool `json:"discoverRedis,omitempty"`
	DiscoverCosmosDB *bool `json:"discoverCosmosDB,omitempty"`
}