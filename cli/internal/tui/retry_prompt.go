package tui
import (
	"time"
	"github.com/clidey/whodb/cli/pkg/styles"
)
type RetryPromptResult struct {
	Timeout time.Duration
	Save    bool
}
type RetryPrompt struct {
	active        bool
	timedOutQuery string
	autoRetried   bool
}
func (r *RetryPrompt) Show(query string) {
	r.active = true
	r.timedOutQuery = query
}
func (r *RetryPrompt) IsActive() bool {
	return r.active
}
func (r *RetryPrompt) TimedOutQuery() string {
	return r.timedOutQuery
}
func (r *RetryPrompt) AutoRetried() bool {
	return r.autoRetried
}
func (r *RetryPrompt) SetAutoRetried(v bool) {
	r.autoRetried = v
}
func (r *RetryPrompt) HandleKeyMsg(keyStr string) (*RetryPromptResult, bool) {
	if !r.active {
		return nil, false
	}
	switch keyStr {
	case "1":
		r.active = false
		return &RetryPromptResult{Timeout: 60 * time.Second, Save: true}, true
	case "2":
		r.active = false
		return &RetryPromptResult{Timeout: 2 * time.Minute, Save: true}, true
	case "3":
		r.active = false
		return &RetryPromptResult{Timeout: 5 * time.Minute, Save: true}, true
	case "4":
		r.active = false
		return &RetryPromptResult{Timeout: 24 * time.Hour, Save: false}, true
	case "esc":
		r.active = false
		r.timedOutQuery = ""
		return nil, true
	}
	return nil, true
}
func (r *RetryPrompt) View() string {
	var b string
	b += styles.RenderErr("Request timed out")
	b += "\n\n"
	b += styles.RenderMuted("Retry with longer timeout:")
	b += "\n"
	b += styles.RenderKey("[1]")
	b += styles.RenderMuted(" 60 seconds  ")
	b += styles.RenderKey("[2]")
	b += styles.RenderMuted(" 2 minutes  ")
	b += styles.RenderKey("[3]")
	b += styles.RenderMuted(" 5 minutes  ")
	b += styles.RenderKey("[4]")
	b += styles.RenderMuted(" No limit")
	b += "\n\n"
	b += styles.RenderHelp("esc", "cancel")
	return b
}