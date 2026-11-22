package cmd
import (
	"slices"
	"github.com/clidey/whodb/cli/internal/config"
	connresolver "github.com/clidey/whodb/cli/internal/connections"
	"github.com/clidey/whodb/cli/internal/sourcetypes"
	"github.com/spf13/cobra"
)
func completeConnectionNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resolver, err := connresolver.NewResolver(false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	conns := resolver.List()
	names := make([]string, len(conns))
	for i, c := range conns {
		names[i] = c.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func completeDatabaseTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ids := sourcetypes.IDs()
	synonyms := sourcetypes.Synonyms()
	types := make([]string, 0, len(ids)+len(synonyms))
	types = append(types, ids...)
	for _, synonym := range synonyms {
		types = append(types, synonym)
	}
	return types, cobra.ShellCompDirectiveNoFileComp
}
func completeOutputFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"auto", "table", "plain", "json", "ndjson", "csv"}, cobra.ShellCompDirectiveNoFileComp
}
func completeExportFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"csv", "excel"}, cobra.ShellCompDirectiveNoFileComp
}
func completeAuditFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"table", "json"}, cobra.ShellCompDirectiveNoFileComp
}
func completeShellNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"bash", "zsh", "fish", "powershell"}, cobra.ShellCompDirectiveNoFileComp
}
func completeProfileNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.LoadConfigWithoutSecrets()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	profiles := cfg.GetProfiles()
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func completeMCPTransports(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"stdio", "http"}, cobra.ShellCompDirectiveNoFileComp
}
func completeMCPSecurityLevels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"strict", "standard", "minimal"}, cobra.ShellCompDirectiveNoFileComp
}
func completeSSLModes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	dbType, _ := cmd.Flags().GetString("type")
	if spec, ok := lookupDatabaseType(dbType); ok {
		modes := make([]string, 0, len(spec.SSLModes))
		for _, mode := range spec.SSLModes {
			modes = append(modes, string(mode.Value))
		}
		return modes, cobra.ShellCompDirectiveNoFileComp
	}
	modeSet := map[string]struct{}{}
	for _, id := range sourcetypes.IDs() {
		for _, mode := range sourcetypes.SSLModes(id) {
			modeSet[string(mode.Value)] = struct{}{}
		}
	}
	modes := make([]string, 0, len(modeSet))
	for mode := range modeSet {
		modes = append(modes, mode)
	}
	slices.Sort(modes)
	return modes, cobra.ShellCompDirectiveNoFileComp
}