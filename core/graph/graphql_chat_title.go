package graph
import (
	ctx "context"
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/baml_client"
	"github.com/clidey/whodb/core/graph/model"
	"github.com/clidey/whodb/core/src/bamlconfig"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/envconfig"
	"github.com/clidey/whodb/core/src/log"
)
type GenerateChatTitleFunc func(c ctx.Context, input model.GenerateChatTitleInput) (*model.GenerateChatTitleResponse, error)
var registeredGenerateChatTitle GenerateChatTitleFunc
func RegisterGenerateChatTitle(fn GenerateChatTitleFunc) {
	registeredGenerateChatTitle = fn
}
func init() {
	registeredGenerateChatTitle = ceGenerateChatTitle
}
func generateChatTitleImpl(c ctx.Context, input model.GenerateChatTitleInput) (*model.GenerateChatTitleResponse, error) {
	if registeredGenerateChatTitle != nil {
		return registeredGenerateChatTitle(c, input)
	}
	return nil, fmt.Errorf("chat title generation not available")
}
func ceGenerateChatTitle(c ctx.Context, input model.GenerateChatTitleInput) (*model.GenerateChatTitleResponse, error) {
	log.Debugf("Generate Chat Title: Started with query=%s, model=%s", input.Query, input.Model)
	query := strings.TrimSpace(input.Query)
	queryLower := strings.ToLower(query)
	if len(query) <= 3 ||
		queryLower == "hi" ||
		queryLower == "hello" ||
		queryLower == "hey" ||
		queryLower == "hi!" ||
		queryLower == "hello!" ||
		queryLower == "hey!" ||
		queryLower == "blah" ||
		queryLower == "test" {
		return &model.GenerateChatTitleResponse{
			Title: "",
		}, nil
	}
	providerId := ""
	if input.ProviderID != nil {
		providerId = *input.ProviderID
	}
	requestToken := ""
	if input.Token != nil {
		requestToken = *input.Token
	}
	requestEndpoint := ""
	if input.Endpoint != nil {
		requestEndpoint = *input.Endpoint
	}
	creds := envconfig.ResolveProviderCredentials(providerId, requestToken, requestEndpoint, input.ModelType)
	externalModel := &engine.ExternalModel{
		Type:     creds.ModelType,
		Token:    creds.Token,
		Model:    input.Model,
		Endpoint: creds.Endpoint,
	}
	titlePrompt := fmt.Sprintf(
		"Generate a very short title (maximum 5 words) for a chat conversation that starts with this question: \"%s\"\n\n"+
			"Return ONLY the title text, nothing else. No quotes, no explanations. The title should be concise and descriptive.",
		input.Query,
	)
	log.Debugf("Generate Chat Title: Calling BAML GenerateChatTitle")
	callOpts := bamlconfig.SetupAIClient(externalModel)
	stream, err := baml_client.Stream.GenerateChatTitle(c, titlePrompt, callOpts...)
	if err != nil {
		log.Debugf("Generate Chat Title: BAML call failed: %v", err)
		return nil, fmt.Errorf("failed to generate title: %w", err)
	}
	var title string
	for chunk := range stream {
		if chunk.IsError {
			log.Debugf("Generate Chat Title: Stream error: %v", chunk.Error)
			return nil, fmt.Errorf("failed to generate title: %w", chunk.Error)
		}
		if chunk.IsFinal {
			finalTitle := chunk.Final()
			if finalTitle != nil {
				title = *finalTitle
			}
			break
		}
	}
	title = strings.Trim(title, `"'`)
	title = strings.TrimSpace(title)
	if len(title) > 50 {
		title = title[:50]
	}
	log.Debugf("Generate Chat Title: Generated title=%s", title)
	return &model.GenerateChatTitleResponse{
		Title: title,
	}, nil
}