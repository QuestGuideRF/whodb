package main
import (
	cliapp "github.com/clidey/whodb/cli/app"
	"github.com/clidey/whodb/cli/internal/bootstrap"
	"github.com/clidey/whodb/cli/pkg/identity"
)
func main() {
	cliapp.Run(cliapp.Config{
		Identity:  identity.CE(),
		Bootstrap: bootstrap.Ensure,
	})
}