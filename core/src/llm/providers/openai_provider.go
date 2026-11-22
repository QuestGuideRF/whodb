package providers
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"github.com/clidey/whodb/core/src/log"
)
var responsesAPICache sync.Map
var httpClientFactory = func() *http.Client { return &http.Client{} }
const (
	OpenAI_LLMType LLMType = "OpenAI"
)
type OpenAIProvider struct{}
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{}
}
func (p *OpenAIProvider) GetType() LLMType {
	return OpenAI_LLMType
}
func (p *OpenAIProvider) GetProtocol() string { return "openai" }
func (p *OpenAIProvider) GetDefaultEndpoint() string {
	return "https://api.openai.com/v1"
}
func (p *OpenAIProvider) ValidateConfig(config *ProviderConfig) error {
	if config.APIKey == "" {
		return errors.New("API key is required for OpenAI")
	}
	if config.Endpoint == "" {
		config.Endpoint = p.GetDefaultEndpoint()
	}
	return nil
}
func (p *OpenAIProvider) GetSupportedModels(config *ProviderConfig) ([]string, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/models", config.Endpoint)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", config.APIKey),
		"Content-Type":  "application/json",
	}
	resp, err := sendHTTPRequest("GET", url, nil, headers)
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch models from OpenAI at %s", url)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("OpenAI models endpoint returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("failed to fetch models: %s", string(body))
	}
	var modelsResp struct {
		Models []struct {
			Name string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		log.WithError(err).Error("Failed to decode OpenAI models response")
		return nil, err
	}
	var models []string
	for _, model := range modelsResp.Models {
		name := model.Name
		if isNonChatModel(name) {
			continue
		}
		models = append(models, name)
	}
	return models, nil
}
func (p *OpenAIProvider) CreateBAMLClient(config *ProviderConfig, model string) (string, map[string]any, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = p.GetDefaultEndpoint()
	}
	clientType := "openai"
	if p.supportsResponsesAPI(endpoint, config.APIKey) {
		clientType = "openai-responses"
	}
	opts := map[string]any{"model": model}
	if config.APIKey != "" {
		opts["api_key"] = config.APIKey
	}
	if config.Endpoint != "" && config.Endpoint != p.GetDefaultEndpoint() {
		opts["base_url"] = config.Endpoint
	}
	return clientType, opts, nil
}
func (p *OpenAIProvider) supportsResponsesAPI(endpoint, apiKey string) bool {
	if cached, ok := responsesAPICache.Load(endpoint); ok {
		return cached.(bool)
	}
	probeURL := endpoint + "/responses"
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if apiKey != "" {
		headers["Authorization"] = "Bearer " + apiKey
	}
	resp, err := sendHTTPRequest("POST", probeURL, []byte("{}"), headers)
	if err != nil {
		log.Debugf("Responses API probe failed for %s (network error), falling back to Chat Completions", endpoint)
		return false
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	supports := resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed
	if supports {
		log.Debugf("Responses API available at %s (probe status: %d)", endpoint, resp.StatusCode)
	} else {
		log.Debugf("Responses API not available at %s (probe status: %d), using Chat Completions", endpoint, resp.StatusCode)
	}
	responsesAPICache.Store(endpoint, supports)
	return supports
}
func isNonChatModel(name string) bool {
	nonChatPrefixes := []string{
		"dall-e-",
		"whisper-",
		"tts-",
		"text-embedding-",
		"text-moderation-",
		"gpt-image-",
		"babbage-",
		"davinci-",
	}
	for _, prefix := range nonChatPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	nonChatPatterns := []string{"-instruct", "-base", "-codex"}
	for _, pattern := range nonChatPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}
	if strings.Contains(name, "-transcribe") ||
		strings.Contains(name, "-realtime") {
		return true
	}
	return false
}
func sendHTTPRequest(method, url string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return httpClientFactory().Do(req)
}