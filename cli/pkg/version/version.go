package version
import (
	"fmt"
	"runtime"
	"strings"
	"github.com/clidey/whodb/cli/pkg/identity"
)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)
type Info struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
	Platform  string
}
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
func (i Info) String() string {
	name := identity.Current().VersionName
	if name == "" {
		name = identity.Current().CommandName
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s\n", name, i.Version))
	b.WriteString(fmt.Sprintf("  Commit:     %s\n", i.Commit))
	b.WriteString(fmt.Sprintf("  Built:      %s\n", i.BuildDate))
	b.WriteString(fmt.Sprintf("  Go version: %s\n", i.GoVersion))
	b.WriteString(fmt.Sprintf("  Platform:   %s", i.Platform))
	return b.String()
}
func Short() string {
	name := identity.Current().VersionName
	if name == "" {
		name = identity.Current().CommandName
	}
	return fmt.Sprintf("%s %s (%s)", name, Version, Commit)
}