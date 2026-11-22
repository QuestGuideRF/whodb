package llm
import (
	"github.com/clidey/whodb/core/src/source"
)
func ClientForModel(model *source.ExternalModel) *LLMClient {
	if model == nil {
		return &LLMClient{}
	}
	return &LLMClient{
		Type:   LLMType(model.Type),
		APIKey: model.Token,
	}
}