package graph
import "net/http"
type AgentHandlerFunc func(w http.ResponseWriter, r *http.Request)
var registeredAgentHandler AgentHandlerFunc
var registeredAgentPermitHandler AgentHandlerFunc
func RegisterAgentHandler(handler AgentHandlerFunc) {
	registeredAgentHandler = handler
}
func RegisterAgentPermitHandler(handler AgentHandlerFunc) {
	registeredAgentPermitHandler = handler
}
func agentStreamHandler(w http.ResponseWriter, r *http.Request) {
	if registeredAgentHandler != nil {
		registeredAgentHandler(w, r)
		return
	}
	http.Error(w, "Agent not available in this edition", http.StatusNotImplemented)
}
func agentPermitHandler(w http.ResponseWriter, r *http.Request) {
	if registeredAgentPermitHandler != nil {
		registeredAgentPermitHandler(w, r)
		return
	}
	http.Error(w, "Agent not available in this edition", http.StatusNotImplemented)
}