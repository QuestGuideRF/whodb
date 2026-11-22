package gcp
import (
	"net"
	"sync"
	"sync/atomic"
	"time"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/providers"
)
const (
	connectivityTimeout     = 3 * time.Second
	maxConcurrentChecks     = 10
	connectivityReachable   = "reachable"
	connectivityUnreachable = "unreachable"
)
func CheckConnectivity(connections []providers.DiscoveredConnection) {
	log.Infof("Connectivity: checking %d connections", len(connections))
	sem := make(chan struct{}, maxConcurrentChecks)
	var wg sync.WaitGroup
	var reachable, unreachable atomic.Int32
	for i := range connections {
		conn := &connections[i]
		endpoint := conn.Metadata["endpoint"]
		port := conn.Metadata["port"]
		if endpoint == "" || port == "" {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			addr := net.JoinHostPort(endpoint, port)
			c, err := net.DialTimeout("tcp", addr, connectivityTimeout)
			if err != nil {
				log.Debugf("Connectivity: %s unreachable: %v", addr, err)
				conn.Metadata["connectivity"] = connectivityUnreachable
				unreachable.Add(1)
				return
			}
			c.Close()
			log.Debugf("Connectivity: %s reachable", addr)
			conn.Metadata["connectivity"] = connectivityReachable
			reachable.Add(1)
		}()
	}
	wg.Wait()
	log.Infof("Connectivity: done — %d reachable, %d unreachable", reachable.Load(), unreachable.Load())
}