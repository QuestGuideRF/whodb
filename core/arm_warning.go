package main
import (
	"fmt"
	"os"
)
func init() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "⚠️  Notice: AI features and DuckDB are unavailable on this architecture (arm/riscv64).")
	fmt.Fprintln(os.Stderr, "   All other database features work normally.")
	fmt.Fprintln(os.Stderr, "")
}