package sourcetypes
import (
	"testing"
	"github.com/clidey/whodb/core/src/source"
)
func TestExplainMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  source.QueryExplainMode
		ok    bool
	}{
		{input: "Postgres", want: source.QueryExplainModeExplainAnalyze, ok: true},
		{input: "ClickHouse", want: source.QueryExplainModeExplainPipeline, ok: true},
		{input: "MongoDB", want: source.QueryExplainModeNone, ok: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, ok := ExplainMode(tt.input)
			if ok != tt.ok {
				t.Fatalf("expected ExplainMode(%q) ok=%t, got %t", tt.input, tt.ok, ok)
			}
			if got != tt.want {
				t.Fatalf("expected ExplainMode(%q)=%q, got %q", tt.input, tt.want, got)
			}
		})
	}
}