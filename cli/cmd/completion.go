package cmd
import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/clidey/whodb/cli/pkg/identity"
	"github.com/spf13/cobra"
)
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate autocompletion scripts for supported shells.
Examples:
  # Install completion for your current shell (recommended)
  whodb-cli completion install
  # Print completion script for bash to stdout
  whodb-cli completion bash|zsh|fish|powershell
  # Install completion for a specific shell
  whodb-cli completion install bash`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		sh := strings.ToLower(strings.TrimSpace(args[0]))
		var buf bytes.Buffer
		if err := generateCompletion(sh, &buf); err != nil {
			return err
		}
		_, _ = io.Copy(os.Stdout, &buf)
		return nil
	},
}
var completionInstallCmd = &cobra.Command{
	Use:   "install [bash|zsh|fish|powershell]",
	Short: "Install shell completion script",
	Long: `Install autocompletion for your shell into standard user directories.
This writes the completion script to a common per-user location without requiring sudo.
For bash and zsh, ensure your shell loads completions from these locations as noted below.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sh := ""
		if len(args) == 1 {
			sh = strings.ToLower(strings.TrimSpace(args[0]))
		} else {
			sh = detectShell()
		}
		if sh == "" {
			return errors.New("could not detect shell; please specify one: bash|zsh|fish|powershell")
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		var targetPath string
		var postMsg string
		switch sh {
		case "bash":
			targetPath = filepath.Join(home, ".local", "share", "bash-completion", "completions", completionCommandName())
			postMsg = "Bash completion installed. Ensure bash-completion is installed and sourced. Restart your shell, or source ~/.bashrc."
		case "zsh":
			targetPath = filepath.Join(home, ".zsh", "completions", "_"+completionCommandName())
			postMsg = "Zsh completion installed. Add 'fpath=(~/.zsh/completions $fpath); autoload -U compinit; compinit' to your ~/.zshrc if not already configured. Then restart zsh."
		case "fish":
			targetPath = filepath.Join(home, ".config", "fish", "completions", completionCommandName()+".fish")
			postMsg = "Fish completion installed. Restart fish or run 'exec fish'."
		case "powershell", "pwsh":
			return fmt.Errorf("powershell completion install is not automated yet; run '%s completion powershell | Out-String | Set-Content -Path $PROFILE' in PowerShell", completionCommandName())
		default:
			return fmt.Errorf("unsupported shell: %s", sh)
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("failed creating directory %s: %w", filepath.Dir(targetPath), err)
		}
		var buf bytes.Buffer
		if err := generateCompletion(sh, &buf); err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, buf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("failed writing completion to %s: %w", targetPath, err)
		}
		fmt.Fprintf(os.Stdout, "Installed %s completion to: %s\n", sh, targetPath)
		switch sh {
		case "bash":
			if err := ensureBashRc(targetPath); err != nil {
				return fmt.Errorf("failed updating bash rc: %w", err)
			}
		case "zsh":
			if err := ensureZshRc(); err != nil {
				return fmt.Errorf("failed updating zsh rc: %w", err)
			}
		case "fish":
		}
		fmt.Fprintln(os.Stdout, postMsg)
		return nil
	},
}
func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.AddCommand(completionInstallCmd)
	completionCmd.AddCommand(completionUninstallCmd)
	completionCmd.ValidArgsFunction = completeShellNames
	completionInstallCmd.ValidArgsFunction = completeShellNames
	completionUninstallCmd.ValidArgsFunction = completeShellNames
}
func detectShell() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		base := filepath.Base(sh)
		base = strings.ToLower(base)
		switch base {
		case "bash", "zsh", "fish":
			return base
		}
	}
	if runtime.GOOS == "windows" {
		if ps := os.Getenv("PSModulePath"); ps != "" {
			return "powershell"
		}
	}
	return ""
}
func generateCompletion(shell string, w io.Writer) error {
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(w)
	case "zsh":
		return rootCmd.GenZshCompletion(w)
	case "fish":
		return rootCmd.GenFishCompletion(w, true)
	case "powershell", "pwsh":
		return rootCmd.GenPowerShellCompletion(w)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}
func ensureBashRc(installPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	block := completionRCBlockStart() + "\n" +
		"# shell completion\n" +
		"if [ -f '" + installPath + "' ]; then\n" +
		". '" + installPath + "'\n" +
		"fi\n" +
		completionRCBlockEnd() + "\n"
	var rcFiles []string
	if runtime.GOOS == "darwin" {
		rcFiles = []string{filepath.Join(home, ".bash_profile"), filepath.Join(home, ".bashrc")}
	} else {
		rcFiles = []string{filepath.Join(home, ".bashrc")}
	}
	for _, rc := range rcFiles {
		if err := ensureBlockInFile(rc, block); err != nil {
			return err
		}
	}
	return nil
}
func ensureZshRc() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	block := completionRCBlockStart() + "\n" +
		"fpath=(~/.zsh/completions $fpath)\n" +
		"autoload -U compinit\n" +
		"compinit\n" +
		completionRCBlockEnd() + "\n"
	rc := filepath.Join(home, ".zshrc")
	return ensureBlockInFile(rc, block)
}
func ensureBlockInFile(path string, block string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			return os.WriteFile(path, []byte(block), 0o644)
		}
		return err
	}
	content := string(data)
	if strings.Contains(content, completionRCBlockStart()) {
		return nil
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += block
	return os.WriteFile(path, []byte(content), 0o644)
}
var completionUninstallCmd = &cobra.Command{
	Use:   "uninstall [bash|zsh|fish|powershell]",
	Short: "Uninstall shell completion script",
	Long:  "Remove installed completion script and shell init blocks. If shell is not provided, the system default will be used.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sh := ""
		if len(args) == 1 {
			sh = strings.ToLower(strings.TrimSpace(args[0]))
		} else {
			sh = detectShell()
		}
		if sh == "" {
			return errors.New("could not detect shell; please specify one: bash|zsh|fish|powershell")
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		var targetPath string
		switch sh {
		case "bash":
			targetPath = filepath.Join(home, ".local", "share", "bash-completion", "completions", completionCommandName())
			if err := removeBashRc(); err != nil {
				return err
			}
		case "zsh":
			targetPath = filepath.Join(home, ".zsh", "completions", "_"+completionCommandName())
			if err := removeZshRc(); err != nil {
				return err
			}
		case "fish":
			targetPath = filepath.Join(home, ".config", "fish", "completions", completionCommandName()+".fish")
		case "powershell", "pwsh":
			fmt.Fprintln(os.Stdout, "PowerShell uninstall is manual: remove completion lines from $PROFILE and any saved completion file.")
			return nil
		default:
			return fmt.Errorf("unsupported shell: %s", sh)
		}
		if err := removeFileIfExists(targetPath); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Removed %s completion from: %s\n", sh, targetPath)
		return nil
	},
}
func removeBashRc() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	var rcFiles []string
	if runtime.GOOS == "darwin" {
		rcFiles = []string{filepath.Join(home, ".bash_profile"), filepath.Join(home, ".bashrc")}
	} else {
		rcFiles = []string{filepath.Join(home, ".bashrc")}
	}
	for _, rc := range rcFiles {
		if err := removeBlockFromFile(rc); err != nil {
			return err
		}
	}
	return nil
}
func removeZshRc() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	rc := filepath.Join(home, ".zshrc")
	return removeBlockFromFile(rc)
}
func removeBlockFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	content := string(data)
	start := strings.Index(content, completionRCBlockStart())
	if start == -1 {
		return nil
	}
	end := strings.Index(content[start:], completionRCBlockEnd())
	if end == -1 {
		return nil
	}
	end += start + len(completionRCBlockEnd())
	newContent := content[:start] + content[end:]
	newContent = strings.ReplaceAll(newContent, "\n\n\n", "\n\n")
	return os.WriteFile(path, []byte(newContent), 0o644)
}
func removeFileIfExists(path string) error {
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}
func completionCommandName() string {
	return identity.Current().CommandName
}
func completionRCBlockStart() string {
	return "# >>> " + completionCommandName() + " completion >>>"
}
func completionRCBlockEnd() string {
	return "# <<< " + completionCommandName() + " completion <<<"
}