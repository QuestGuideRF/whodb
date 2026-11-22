package aws
import (
	"context"
	"time"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awsinfra "github.com/clidey/whodb/core/src/aws"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/providers"
)
func (p *Provider) discoverOpenSearch(ctx context.Context) ([]providers.DiscoveredConnection, error) {
	start := time.Now()
	log.Debugf("OpenSearch: starting discovery for provider %s", p.config.ID)
	client := opensearch.NewFromConfig(p.awsConfig)
	listOutput, err := client.ListDomainNames(ctx, &opensearch.ListDomainNamesInput{})
	if err != nil {
		log.Errorf("OpenSearch: ListDomainNames failed: %v", err)
		return nil, awsinfra.HandleAWSError(err)
	}
	if len(listOutput.DomainNames) == 0 {
		log.Debugf("OpenSearch: no domains found")
		return nil, nil
	}
	domainNames := make([]string, 0, len(listOutput.DomainNames))
	for _, d := range listOutput.DomainNames {
		if d.DomainName != nil {
			domainNames = append(domainNames, *d.DomainName)
		}
	}
	descOutput, err := client.DescribeDomains(ctx, &opensearch.DescribeDomainsInput{
		DomainNames: domainNames,
	})
	if err != nil {
		log.Errorf("OpenSearch: DescribeDomains failed: %v", err)
		return nil, awsinfra.HandleAWSError(err)
	}
	var connections []providers.DiscoveredConnection
	for _, domain := range descOutput.DomainStatusList {
		if domain.DomainName == nil {
			continue
		}
		name := *domain.DomainName
		metadata := map[string]string{
			"port": "443",
		}
		if domain.Endpoint != nil {
			metadata["endpoint"] = *domain.Endpoint
		} else if domain.EndpointV2 != nil {
			metadata["endpoint"] = *domain.EndpointV2
		}
		if domain.EngineVersion != nil {
			metadata["version"] = *domain.EngineVersion
		}
		status := mapOpenSearchStatus(domain.Created, domain.Deleted, aws.ToBool(domain.Processing))
		conn := providers.DiscoveredConnection{
			ID:           p.connectionID("opensearch-" + name),
			ProviderType: providers.ProviderTypeAWS,
			ProviderID:   p.config.ID,
			Name:         name,
			DatabaseType: engine.DatabaseType_OpenSearch,
			Region:       p.config.Region,
			Status:       status,
			Metadata:     metadata,
		}
		connections = append(connections, conn)
	}
	log.Infof("OpenSearch: discovered %d domains in %v", len(connections), time.Since(start))
	return connections, nil
}
func mapOpenSearchStatus(created, deleted *bool, processing bool) providers.ConnectionStatus {
	if deleted != nil && *deleted {
		return providers.ConnectionStatusDeleting
	}
	if created != nil && !*created {
		return providers.ConnectionStatusStarting
	}
	if processing {
		return providers.ConnectionStatusStarting
	}
	return providers.ConnectionStatusAvailable
}