package app
import (
	"github.com/clidey/whodb/cli/cmd"
	"github.com/clidey/whodb/cli/pkg/crash"
	"github.com/clidey/whodb/cli/pkg/identity"
)
type Config struct {
	Identity  identity.Config
	Bootstrap func()
}
func Run(config Config) {
	identity.SetCurrent(config.Identity)
	if config.Bootstrap != nil {
		config.Bootstrap()
	}
	defer crash.Handler()
	cmd.Execute()
}