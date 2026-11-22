package source
import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
)
var (
	driversMu sync.RWMutex
	drivers   = map[string]SourceConnector{}
	typesMu       sync.RWMutex
	registeredIDs []string
	registered    = map[string]TypeSpec{}
)
func RegisterDriver(id string, connector SourceConnector) {
	if id == "" || connector == nil {
		return
	}
	driversMu.Lock()
	defer driversMu.Unlock()
	drivers[id] = connector
}
func Open(ctx context.Context, spec TypeSpec, credentials *Credentials) (SourceSession, error) {
	driversMu.RLock()
	driver, ok := drivers[spec.DriverID]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unsupported source driver: %s", spec.DriverID)
	}
	return driver.Open(ctx, spec, credentials)
}
func Invalidate(ctx context.Context, spec TypeSpec, credentials *Credentials) error {
	driversMu.RLock()
	driver, ok := drivers[spec.DriverID]
	driversMu.RUnlock()
	if !ok {
		return fmt.Errorf("unsupported source driver: %s", spec.DriverID)
	}
	invalidator, ok := driver.(SessionInvalidator)
	if !ok {
		return nil
	}
	return invalidator.Invalidate(ctx, spec, credentials)
}
func Shutdown(ctx context.Context) error {
	driversMu.RLock()
	driverList := make([]SourceConnector, 0, len(drivers))
	for _, driver := range drivers {
		driverList = append(driverList, driver)
	}
	driversMu.RUnlock()
	var shutdownErr error
	for _, driver := range driverList {
		shutdowner, ok := driver.(DriverShutdowner)
		if !ok {
			continue
		}
		shutdownErr = errors.Join(shutdownErr, shutdowner.Shutdown(ctx))
	}
	return shutdownErr
}
func RegisterType(spec TypeSpec) {
	if spec.ID == "" {
		return
	}
	typesMu.Lock()
	defer typesMu.Unlock()
	key := strings.ToLower(spec.ID)
	if _, exists := registered[key]; !exists {
		registeredIDs = append(registeredIDs, key)
	}
	registered[key] = cloneTypeSpec(spec)
}
func RegisteredTypes() []TypeSpec {
	typesMu.RLock()
	defer typesMu.RUnlock()
	specs := make([]TypeSpec, 0, len(registeredIDs))
	for _, key := range registeredIDs {
		spec, ok := registered[key]
		if !ok {
			continue
		}
		specs = append(specs, cloneTypeSpec(spec))
	}
	return specs
}
func FindType(id string) (TypeSpec, bool) {
	typesMu.RLock()
	defer typesMu.RUnlock()
	spec, ok := registered[strings.ToLower(id)]
	if !ok {
		return TypeSpec{}, false
	}
	return cloneTypeSpec(spec), true
}
func cloneTypeSpec(spec TypeSpec) TypeSpec {
	cloned := spec
	cloned.ConnectionFields = slices.Clone(spec.ConnectionFields)
	cloned.Contract = Contract{
		Model:             spec.Contract.Model,
		Surfaces:          slices.Clone(spec.Contract.Surfaces),
		RootActions:       slices.Clone(spec.Contract.RootActions),
		BrowsePath:        slices.Clone(spec.Contract.BrowsePath),
		DefaultObjectKind: spec.Contract.DefaultObjectKind,
		GraphScopeKind:    spec.Contract.GraphScopeKind,
		ObjectTypes:       cloneObjectTypes(spec.Contract.ObjectTypes),
	}
	cloned.DiscoveryPrefill = cloneDiscoveryPrefill(spec.DiscoveryPrefill)
	cloned.SSLModes = cloneSSLModes(spec.SSLModes)
	return cloned
}
func cloneObjectTypes(objectTypes []ObjectType) []ObjectType {
	cloned := make([]ObjectType, 0, len(objectTypes))
	for _, objectType := range objectTypes {
		cloned = append(cloned, ObjectType{
			Kind:          objectType.Kind,
			DataShape:     objectType.DataShape,
			Actions:       slices.Clone(objectType.Actions),
			Views:         slices.Clone(objectType.Views),
			SingularLabel: objectType.SingularLabel,
			PluralLabel:   objectType.PluralLabel,
		})
	}
	return cloned
}
func cloneDiscoveryPrefill(prefill DiscoveryPrefill) DiscoveryPrefill {
	cloned := DiscoveryPrefill{
		AdvancedDefaults: make([]DiscoveryAdvancedDefault, 0, len(prefill.AdvancedDefaults)),
	}
	for _, item := range prefill.AdvancedDefaults {
		cloned.AdvancedDefaults = append(cloned.AdvancedDefaults, DiscoveryAdvancedDefault{
			Key:           item.Key,
			Value:         item.Value,
			MetadataKey:   item.MetadataKey,
			DefaultValue:  item.DefaultValue,
			ProviderTypes: slices.Clone(item.ProviderTypes),
			Conditions:    slices.Clone(item.Conditions),
		})
	}
	return cloned
}
func cloneSSLModes(modes []SSLModeInfo) []SSLModeInfo {
	cloned := make([]SSLModeInfo, 0, len(modes))
	for _, mode := range modes {
		cloned = append(cloned, SSLModeInfo{
			Value:       mode.Value,
			Label:       mode.Label,
			Description: mode.Description,
			Aliases:     slices.Clone(mode.Aliases),
		})
	}
	return cloned
}