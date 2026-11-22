package crash
import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"github.com/clidey/whodb/cli/pkg/identity"
	"github.com/clidey/whodb/cli/pkg/version"
)
func Handler() {
	if r := recover(); r != nil {
		printCrashReport(r)
		os.Exit(1)
	}
}
func printCrashReport(err any) {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])
	v := version.Get()
	cmdLine := strings.Join(os.Args, " ")
	cfg := identity.Current()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "============================================================")
	fmt.Fprintf(os.Stderr, "  %s crashed unexpectedly!\n", cfg.DisplayName)
	fmt.Fprintln(os.Stderr, "============================================================")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Please report this issue at:")
	fmt.Fprintln(os.Stderr, "  "+cfg.IssueURL)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Copy and paste the following into the issue:")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "------------------------------------------------------------")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, `**Describe the bug**
%s crashed with an unexpected error.
**Error**
%v
**To Reproduce**
The command that caused the crash:
`+"```"+`
%s
`+"```"+`
**Expected behavior**
The command should complete without crashing.
**Desktop (please complete the following information):**
- OS: %s
- Architecture: %s
- %s Version: %s
- Commit: %s
- Built: %s
- Go Version: %s
**Stack Trace**
`+"```"+`
%s
`+"```"+`
**Additional context**
[Add any additional context here, such as what you were trying to do]
`,
		cfg.DisplayName,
		err,
		cmdLine,
		runtime.GOOS,
		runtime.GOARCH,
		cfg.DisplayName,
		v.Version,
		v.Commit,
		v.BuildDate,
		v.GoVersion,
		stackTrace,
	)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "------------------------------------------------------------")
}