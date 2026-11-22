package mcp
import (
	"fmt"
	"strings"
	connresolver "github.com/clidey/whodb/cli/internal/connections"
	dbmgr "github.com/clidey/whodb/cli/internal/database"
)
type connectionResolver struct {
	resolver *connresolver.Resolver
}
func newConnectionResolver(includeSecrets bool) (*connectionResolver, error) {
	resolver, err := connresolver.NewResolver(includeSecrets)
	if err != nil {
		return nil, err
	}
	return &connectionResolver{resolver: resolver}, nil
}
func (r *connectionResolver) Count() int {
	if r == nil {
		return 0
	}
	return r.resolver.Count()
}
func (r *connectionResolver) Resolve(name string) (*dbmgr.Connection, error) {
	conn, _, err := r.resolver.Resolve(name)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
func (r *connectionResolver) ResolveOrDefault(name string) (*dbmgr.Connection, error) {
	if strings.TrimSpace(name) != "" {
		return r.Resolve(name)
	}
	infos := r.resolver.ListWithSource()
	switch len(infos) {
	case 0:
		return nil, fmt.Errorf("no database connections available. Add one with: whodb-cli connections add or set WHODB_<DBTYPE> profiles")
	case 1:
		conn := infos[0].Connection
		return &conn, nil
	default:
		names := make([]string, len(infos))
		for i, info := range infos {
			names[i] = info.Connection.Name
		}
		return nil, fmt.Errorf("multiple connections available (%d). Please specify which one to use. Available: %v", len(names), names)
	}
}
func (r *connectionResolver) ListNames() []string {
	if r == nil {
		return nil
	}
	infos := r.resolver.ListWithSource()
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Connection.Name)
	}
	return names
}
func (r *connectionResolver) ListWithSource() []connresolver.ConnectionSourceInfo {
	if r == nil {
		return nil
	}
	return r.resolver.ListWithSource()
}
func ResolveConnection(name string) (*dbmgr.Connection, error) {
	resolver, err := newConnectionResolver(true)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize database manager: %w", err)
	}
	return resolver.Resolve(name)
}
func ListAvailableConnections() ([]string, error) {
	resolver, err := newConnectionResolver(false)
	if err != nil {
		return nil, err
	}
	return resolver.ListNames(), nil
}
func ResolveConnectionOrDefault(name string) (*dbmgr.Connection, error) {
	resolver, err := newConnectionResolver(true)
	if err != nil {
		return nil, err
	}
	return resolver.ResolveOrDefault(name)
}