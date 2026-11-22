package graph
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"github.com/clidey/whodb/core/graph/model"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/source"
	"golang.org/x/sync/errgroup"
)
type StreamRequest struct {
	ModelType  string                      `json:"modelType"`
	Token      string                      `json:"token"`
	Model      string                      `json:"model"`
	Endpoint   string                      `json:"endpoint"`
	Ref        *model.SourceObjectRefInput `json:"ref"`
	ProviderId string                      `json:"providerId"`
	Input      model.ChatInput             `json:"input"`
}
func ParseStreamRequest(r *http.Request) (*StreamRequest, error) {
	var req StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}
func SetupSSEHeaders(w http.ResponseWriter) http.Flusher {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher, _ := w.(http.Flusher)
	return flusher
}
func BuildObjectDetails(ctx context.Context, session source.SourceSession, parent *source.ObjectRef, kind source.ObjectKind) (string, error) {
	browser, ok := session.(source.SourceBrowser)
	if !ok {
		return "", errors.New("source browsing is not supported")
	}
	reader, ok := session.(source.TabularReader)
	if !ok {
		return "", errors.New("source rows are not supported")
	}
	objects, err := browser.ListObjects(ctx, parent, []source.ObjectKind{kind})
	if err != nil {
		return "", err
	}
	type objectResult struct {
		name    string
		kind    source.ObjectKind
		columns []source.Column
	}
	results := make([]objectResult, len(objects))
	var mu sync.Mutex
	g := new(errgroup.Group)
	g.SetLimit(10)
	for i, object := range objects {
		i, object := i, object
		g.Go(func() error {
			columns, err := reader.Columns(ctx, object.Ref)
			if err != nil {
				log.WithError(err).Warnf("Failed to get columns for %s %s in streaming chat", object.Kind, object.Name)
				return nil
			}
			mu.Lock()
			results[i] = objectResult{name: object.Name, kind: object.Kind, columns: columns}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return "", err
	}
	var b strings.Builder
	for _, r := range results {
		if r.name == "" {
			continue
		}
		fmt.Fprintf(&b, "%s: %s\n", strings.ToLower(string(r.kind)), r.name)
		for _, col := range r.columns {
			fmt.Fprintf(&b, "- %s (%s)\n", col.Name, col.Type)
		}
	}
	return b.String(), nil
}
func SendSSEMessage(w http.ResponseWriter, flusher http.Flusher, message *model.AIChatMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("Failed to marshal SSE message")
		return
	}
	fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
	flusher.Flush()
}
func SendSSEChunk(w http.ResponseWriter, flusher http.Flusher, chunk map[string]any) {
	data, err := json.Marshal(chunk)
	if err != nil {
		log.WithError(err).Error("Failed to marshal SSE chunk")
		return
	}
	fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", data)
	flusher.Flush()
}
func SendSSEError(w http.ResponseWriter, flusher http.Flusher, errorMsg string) {
	sanitized := sanitizeErrorMessage(errorMsg)
	SendSSEMessage(w, flusher, &model.AIChatMessage{
		Type: "error",
		Text: sanitized,
	})
	SendSSEDone(w, flusher)
}
func sanitizeErrorMessage(msg string) string {
	if strings.Contains(msg, "LLM Failure:") || strings.Contains(msg, "ServiceError") {
		if strings.Contains(msg, "throttling") || strings.Contains(msg, "rate limit") {
			return "AI service is busy. Please try again in a moment."
		}
		if strings.Contains(msg, "access denied") || strings.Contains(msg, "AccessDenied") {
			return "Access denied. Please check your credentials."
		}
		if strings.Contains(msg, "validation") || strings.Contains(msg, "ValidationException") {
			return "Invalid request. Please check your model selection."
		}
		return "AI service error. Please try again."
	}
	if strings.Contains(msg, "reqwest::Error") || strings.Contains(msg, "Failed to build request") {
		return "Unable to connect to AI service. Please check your configuration."
	}
	if strings.Contains(msg, "RelativeUrlWithoutBase") {
		return "AI service configuration error. Please contact support."
	}
	if strings.Contains(msg, "Failed to start stream") || strings.Contains(msg, "Failed to stream") {
		return "Unable to query. Please try again."
	}
	if strings.Contains(msg, "API key") || strings.Contains(msg, "api_key") ||
		strings.Contains(msg, "OPENAI_API_KEY") || strings.Contains(msg, "unauthorized") {
		return "AI service not configured. Please set up your API key."
	}
	if strings.Contains(msg, "model") && (strings.Contains(msg, "not found") || strings.Contains(msg, "invalid")) {
		return "Selected model is not available. Please choose a different model."
	}
	if len(msg) > 150 || strings.Contains(msg, "{") || strings.Contains(msg, "---") ||
		strings.Contains(msg, "Error {") || strings.Contains(msg, "Prompt:") {
		for _, prefix := range []string{"Failed to", "Unable to", "Error:", "error:"} {
			if idx := strings.Index(msg, prefix); idx >= 0 {
				remaining := msg[idx:]
				for _, end := range []string{".", ":", "\n"} {
					if endIdx := strings.Index(remaining, end); endIdx > 10 && endIdx < 100 {
						return remaining[:endIdx] + "."
					}
				}
			}
		}
		return "Unable to query. Please try again."
	}
	return msg
}
func SendSSEDone(w http.ResponseWriter, flusher http.Flusher) {
	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}
func ConvertResultToMessage(result *source.RowsResult) *model.RowsResult {
	if result == nil {
		return nil
	}
	var columns []*model.Column
	for _, column := range result.Columns {
		columns = append(columns, &model.Column{
			Type: column.Type,
			Name: column.Name,
		})
	}
	return &model.RowsResult{
		Columns: columns,
		Rows:    result.Rows,
	}
}