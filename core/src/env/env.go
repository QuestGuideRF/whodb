package env
import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"github.com/clidey/whodb/core/src/common"
)
var IsDevelopment = os.Getenv("ENVIRONMENT") == "dev"
var IsEnterpriseEdition = false
var ActiveDatabases []string
func GetIsDesktopMode() bool {
	return os.Getenv("WHODB_DESKTOP") == "true"
}
func GetIsCLIMode() bool {
	return os.Getenv("WHODB_CLI") == "true"
}
func GetIsLocalMode() bool {
	return GetIsDesktopMode() || GetIsCLIMode()
}
var Tokens = common.FilterList(strings.Split(os.Getenv("WHODB_TOKENS"), ","), func(item string) bool {
	return item != ""
})
var IsAPIGatewayEnabled = len(Tokens) > 0
var OllamaHost = os.Getenv("WHODB_OLLAMA_HOST")
var OllamaPort = os.Getenv("WHODB_OLLAMA_PORT")
var AnthropicAPIKey = os.Getenv("WHODB_ANTHROPIC_API_KEY")
var AnthropicEndpoint = os.Getenv("WHODB_ANTHROPIC_ENDPOINT")
var AnthropicName = os.Getenv("WHODB_ANTHROPIC_NAME")
var OpenAIAPIKey = os.Getenv("WHODB_OPENAI_API_KEY")
var OpenAIEndpoint = os.Getenv("WHODB_OPENAI_ENDPOINT")
var OpenAIName = os.Getenv("WHODB_OPENAI_NAME")
var OllamaName = os.Getenv("WHODB_OLLAMA_NAME")
var LMStudioBaseURL = os.Getenv("WHODB_LMSTUDIO_BASE_URL")
var LMStudioAPIKey = os.Getenv("WHODB_LMSTUDIO_API_KEY")
var LMStudioName = os.Getenv("WHODB_LMSTUDIO_NAME")
var AllowedOrigins = common.FilterList(strings.Split(os.Getenv("WHODB_ALLOWED_ORIGINS"), ","), func(item string) bool {
	return item != ""
})
var LogLevel = os.Getenv("WHODB_LOG_LEVEL")
var AccessLogFile = os.Getenv("WHODB_ACCESS_LOG_FILE")
var LogFile = os.Getenv("WHODB_LOG_FILE")
var LogFormat = os.Getenv("WHODB_LOG_FORMAT")
var BasePath = getBasePath()
const DefaultLogDir = "/var/log/whodb"
const DefaultLogFile = DefaultLogDir + "/whodb.log"
const DefaultAccessLogFile = DefaultLogDir + "/whodb.access.log"
func GetDisableUpdateCheck() bool {
	return os.Getenv("WHODB_DISABLE_UPDATE_CHECK") == "true"
}
var DisableMockDataGeneration = os.Getenv("WHODB_DISABLE_MOCK_DATA_GENERATION")
var ApplicationEnvironment = os.Getenv("WHODB_APPLICATION_ENVIRONMENT")
var ApplicationVersion string
var PosthogAPIKey = "phc_hbXcCoPTdxm5ADL8PmLSYTIUvS6oRWFM2JAK8SMbfnH"
var PosthogHost = "https://us.i.posthog.com"
var IsAWSProviderEnabled = os.Getenv("WHODB_ENABLE_AWS_PROVIDER") == "true"
var IsAzureProviderEnabled = os.Getenv("WHODB_ENABLE_AZURE_PROVIDER") == "true"
var IsGCPProviderEnabled = os.Getenv("WHODB_ENABLE_GCP_PROVIDER") == "true"
var DisableCredentialForm = os.Getenv("WHODB_DISABLE_CREDENTIAL_FORM") == "true"
var BridgeURL = os.Getenv("WHODB_BRIDGE_URL")
var MaxPageSize = getMaxPageSize()
type ChatProvider struct {
	Type       string
	Name       string
	APIKey     string
	Endpoint   string
	ProviderId string
	ClientType string
	IsGeneric  bool
}
type GenericProviderConfig struct {
	ProviderId string
	Name       string
	ClientType string
	BaseURL    string
	APIKey     string
	Models     []string
}
var GenericProviders []GenericProviderConfig
func AddGenericProvider(config GenericProviderConfig) {
	GenericProviders = append(GenericProviders, config)
}
type ResolvedCredentials struct {
	ModelType string
	Token     string
	Endpoint  string
}
func GetOllamaEndpoint() string {
	host := "localhost"
	port := "11434"
	if OllamaHost != "" {
		host = OllamaHost
	}
	if OllamaPort != "" {
		port = OllamaPort
	}
	return common.ResolveLocalURL(fmt.Sprintf("http://%s:%s/api", host, port))
}
func GetAnthropicEndpoint() string {
	if AnthropicEndpoint != "" {
		return AnthropicEndpoint
	}
	return "https://api.anthropic.com/v1"
}
func GetOpenAIEndpoint() string {
	if OpenAIEndpoint != "" {
		return OpenAIEndpoint
	}
	return "https://api.openai.com/v1"
}
func GetLMStudioEndpoint() string {
	if LMStudioBaseURL != "" {
		return common.ResolveLocalURL(LMStudioBaseURL)
	}
	return common.ResolveLocalURL("http://localhost:1234/v1")
}
func getMaxPageSize() int {
	val := os.Getenv("WHODB_MAX_PAGE_SIZE")
	if val == "" {
		return 10000
	}
	n, err := strconv.Atoi(val)
	if err != nil || n <= 0 {
		return 10000
	}
	return n
}
func getBasePath() string {
	basePath := strings.TrimSpace(os.Getenv("WHODB_BASE_PATH"))
	if basePath == "" || basePath == "/" {
		return ""
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	basePath = strings.TrimRight(basePath, "/")
	if basePath == "" || basePath == "/" {
		return ""
	}
	if !isValidBasePath(basePath) {
		panic(fmt.Sprintf("invalid WHODB_BASE_PATH %q: must be a slash-prefixed path made of non-empty segments containing only letters, numbers, '.', '_' or '-'", basePath))
	}
	return basePath
}
func isValidBasePath(basePath string) bool {
	if !strings.HasPrefix(basePath, "/") {
		return false
	}
	segments := strings.Split(basePath, "/")[1:]
	if len(segments) == 0 {
		return false
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
		for _, char := range segment {
			switch {
			case char >= 'a' && char <= 'z':
			case char >= 'A' && char <= 'Z':
			case char >= '0' && char <= '9':
			case char == '.', char == '_', char == '-':
			default:
				return false
			}
		}
	}
	return true
}