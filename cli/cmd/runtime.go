package cmd
import (
	"sync"
	"github.com/clidey/whodb/cli/pkg/identity"
	"github.com/spf13/cobra"
)
type commandText struct {
	Use     string
	Short   string
	Long    string
	Example string
}
var commandTextStore sync.Map
func configureRuntime() {
	cfg := identity.Current()
	configureCommandTree(rootCmd)
	if cfg.RootLongAppend != "" {
		rootCmd.Long += "\n" + cfg.RootLongAppend
	}
}
func configureCommandTree(command *cobra.Command) {
	if command == nil {
		return
	}
	text := loadCommandText(command)
	command.Use = identity.ReplaceText(text.Use)
	command.Short = identity.ReplaceText(text.Short)
	command.Long = identity.ReplaceText(text.Long)
	command.Example = identity.ReplaceText(text.Example)
	for _, child := range command.Commands() {
		configureCommandTree(child)
	}
}
func loadCommandText(command *cobra.Command) commandText {
	if existing, ok := commandTextStore.Load(command); ok {
		return existing.(commandText)
	}
	text := commandText{
		Use:     command.Use,
		Short:   command.Short,
		Long:    command.Long,
		Example: command.Example,
	}
	actual, _ := commandTextStore.LoadOrStore(command, text)
	return actual.(commandText)
}