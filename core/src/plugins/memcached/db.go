package memcached
import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
const minVersion = "1.4.31"
func DB(config *engine.PluginConfig) (*Client, error) {
	port := common.GetRecordValueOrDefault(config.Credentials.Advanced, "Port", "11211")
	if _, err := strconv.Atoi(port); err != nil {
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to parse Memcached port number")
		return nil, err
	}
	addr := net.JoinHostPort(config.Credentials.Hostname, port)
	sslConfig := ssl.ParseSSLConfig(engine.DatabaseType_Memcached, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
	var client *Client
	var err error
	if sslConfig != nil && sslConfig.IsEnabled() {
		tlsConfig, tlsErr := ssl.BuildTLSConfig(sslConfig, config.Credentials.Hostname)
		if tlsErr != nil {
			log.WithError(tlsErr).WithFields(map[string]any{
				"hostname": config.Credentials.Hostname,
				"sslMode":  sslConfig.Mode,
			}).Error("Failed to build TLS configuration for Memcached")
			return nil, tlsErr
		}
		client, err = DialTLS(addr, tlsConfig)
	} else {
		client, err = Dial(addr)
	}
	if err != nil {
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to connect to Memcached")
		return nil, err
	}
	if config.Credentials.Username != "" && config.Credentials.Password != "" {
		if authErr := client.Authenticate(config.Credentials.Username, config.Credentials.Password); authErr != nil {
			client.Close()
			log.WithError(authErr).WithField("hostname", config.Credentials.Hostname).Error("Failed to authenticate with Memcached")
			return nil, authErr
		}
	}
	version, err := client.Version()
	if err != nil {
		client.Close()
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to get Memcached version")
		return nil, err
	}
	if !isVersionSupported(version) {
		client.Close()
		return nil, fmt.Errorf("memcached version %s is not supported, minimum required: %s", version, minVersion)
	}
	return client, nil
}
func isVersionSupported(version string) bool {
	return compareVersions(version, minVersion) >= 0
}
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}
	for i := 0; i < maxLen; i++ {
		var aVal, bVal int
		if i < len(aParts) {
			aVal, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bVal, _ = strconv.Atoi(bParts[i])
		}
		if aVal < bVal {
			return -1
		}
		if aVal > bVal {
			return 1
		}
	}
	return 0
}