package graph
import "net/http"
func agentStreamHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Agent not available on this platform", http.StatusNotImplemented)
}
func agentPermitHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Agent not available on this platform", http.StatusNotImplemented)
}