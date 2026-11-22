package engine
var globalPlugins []*Plugin
var pluginTypeAliases = map[DatabaseType]DatabaseType{}
func RegisterPlugin(p *Plugin) {
	globalPlugins = append(globalPlugins, p)
}
func RegisteredPlugins() []*Plugin {
	return globalPlugins
}
func RegisterPluginTypeAlias(alias DatabaseType, target DatabaseType) {
	pluginTypeAliases[alias] = target
}