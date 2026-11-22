package elasticsearch
import (
	"bytes"
	"io"
	"strings"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)
func formatElasticError(res *esapi.Response) string {
	if res == nil {
		return "unknown error"
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.String()
	}
	res.Body = io.NopCloser(bytes.NewBuffer(body))
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return res.String()
	}
	return msg
}