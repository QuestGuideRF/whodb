package cmd
import (
	"fmt"
	"github.com/clidey/whodb/cli/pkg/version"
	"github.com/spf13/cobra"
)
var versionShort bool
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print detailed version information including build date, commit hash, and Go version.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionShort {
			fmt.Println(version.Short())
		} else {
			fmt.Println(version.Get())
		}
	},
}
func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "print short version")
}