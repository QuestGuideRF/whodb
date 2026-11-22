package database
import (
	"context"
	"fmt"
)
func (m *Manager) SendAIChatStream(_ context.Context, _, _, _, _, _, _, _ string) (<-chan StreamChunk, error) {
	return nil, fmt.Errorf("AI streaming not supported on this platform")
}