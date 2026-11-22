package graph
import "net/http"
type AIChatStreamHandlerFunc func(w http.ResponseWriter, r *http.Request)
var registeredAIChatStreamHandler AIChatStreamHandlerFunc
func RegisterAIChatStreamHandler(handler AIChatStreamHandlerFunc) {
	registeredAIChatStreamHandler = handler
}
func aiChatStreamHandler(w http.ResponseWriter, r *http.Request) {
	if registeredAIChatStreamHandler != nil {
		registeredAIChatStreamHandler(w, r)
		return
	}
	http.Error(w, "AI chat streaming not available", http.StatusNotImplemented)
}