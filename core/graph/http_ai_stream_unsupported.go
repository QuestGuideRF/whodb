package graph
import (
	"encoding/json"
	"fmt"
	"net/http"
)
func aiChatStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	errorData := map[string]string{
		"error": "AI features are not supported on this platform (arm/riscv64). BAML requires amd64 or arm64 architecture.",
	}
	data, _ := json.Marshal(errorData)
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", data)
	flusher.Flush()
}