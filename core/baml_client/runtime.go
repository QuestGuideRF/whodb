package baml_client
import (
	"fmt"
	"os"
	"strings"
	baml "github.com/boundaryml/baml/engine/language_client_go/pkg"
	"github.com/clidey/whodb/core/baml_client/type_builder"
)
var bamlRuntime *baml.BamlRuntime
func getEnvVars(overrides map[string]string) map[string]string {
	var env = map[string]string{}
	for _, env_var := range os.Environ() {
		key, value, _ := strings.Cut(env_var, "=")
		env[key] = value
	}
	for key, value := range overrides {
		if value != "" {
			env[key] = value
		} else {
			delete(env, key)
		}
	}
	return env
}
func init() {
	baml.SetTypeMap(typeMap)
	runtime, err := baml.CreateRuntime("./baml_src", getBamlFiles(), getEnvVars(nil))
	if err != nil {
		panic(err)
	}
	bamlRuntime = &runtime
}
type callOption struct {
	client         *string
	clientRegistry *baml.ClientRegistry
	env            map[string]string
	tags           map[string]string
	collectors     []baml.Collector
	onTick         baml.OnTickCallbackData
	typeBuilder    baml.TypeBuilder
}
type CallOptionFunc func(*callOption)
func WithClient(client string) CallOptionFunc {
	return func(o *callOption) {
		o.client = &client
	}
}
func WithClientRegistry(clientRegistry *baml.ClientRegistry) CallOptionFunc {
	return func(o *callOption) {
		o.clientRegistry = clientRegistry
	}
}
func WithEnv(env map[string]string) CallOptionFunc {
	return func(o *callOption) {
		o.env = env
	}
}
func WithTags(tags map[string]string) CallOptionFunc {
	return func(o *callOption) {
		o.tags = tags
	}
}
func WithCollector(collector baml.Collector) CallOptionFunc {
	return func(o *callOption) {
		if o.collectors == nil {
			o.collectors = []baml.Collector{}
		}
		o.collectors = append(o.collectors, collector)
	}
}
type onTickCallbackData struct {
	collector baml.Collector
	onTick    baml.TickCallback
}
func (o *onTickCallbackData) Collector() baml.Collector {
	return o.collector
}
func (o *onTickCallbackData) OnTick() baml.TickCallback {
	return o.onTick
}
func WithOnTick(onTick baml.TickCallback) CallOptionFunc {
	return func(o *callOption) {
		collector, err := bamlRuntime.NewCollector("on-tick-collector")
		if err != nil {
			panic(err)
		}
		if o.collectors == nil {
			o.collectors = []baml.Collector{}
		}
		o.collectors = append(o.collectors, collector)
		o.onTick = &onTickCallbackData{
			collector: collector,
			onTick:    onTick,
		}
	}
}
func WithExperimentalOnTick(onTick baml.TickCallback) CallOptionFunc {
	fmt.Println("Warning: WithExperimentalOnTick is deprecated. Use WithOnTick instead.")
	return WithOnTick(onTick)
}
func WithCollectors(collectors []baml.Collector) CallOptionFunc {
	return func(o *callOption) {
		if collectors == nil {
			return
		}
		if o.collectors == nil {
			o.collectors = []baml.Collector{}
		}
		o.collectors = append(o.collectors, collectors...)
	}
}
type Collector = baml.Collector
type Image = baml.Image
type Audio = baml.Audio
type PDF = baml.PDF
type Video = baml.Video
func NewCollector(name string) (Collector, error) {
	return bamlRuntime.NewCollector(name)
}
func NewImageFromBase64(base64 string, mimeType *string) (Image, error) {
	return bamlRuntime.NewImageFromBase64(base64, mimeType)
}
func NewAudioFromBase64(base64 string, mimeType *string) (Audio, error) {
	return bamlRuntime.NewAudioFromBase64(base64, mimeType)
}
func NewPDFFromBase64(base64 string, mimeType *string) (PDF, error) {
	return bamlRuntime.NewPDFFromBase64(base64, mimeType)
}
func NewVideoFromBase64(base64 string, mimeType *string) (Video, error) {
	return bamlRuntime.NewVideoFromBase64(base64, mimeType)
}
func NewImageFromUrl(url string, mimeType *string) (Image, error) {
	return bamlRuntime.NewImageFromUrl(url, mimeType)
}
func NewAudioFromUrl(url string, mimeType *string) (Audio, error) {
	return bamlRuntime.NewAudioFromUrl(url, mimeType)
}
func NewPDFFromUrl(url string, mimeType *string) (PDF, error) {
	return bamlRuntime.NewPDFFromUrl(url, mimeType)
}
func NewVideoFromUrl(url string, mimeType *string) (Video, error) {
	return bamlRuntime.NewVideoFromUrl(url, mimeType)
}
type TypeBuilder = type_builder.TypeBuilder
func NewTypeBuilder() (*TypeBuilder, error) {
	tb, err := bamlRuntime.NewTypeBuilder()
	if err != nil {
		return nil, err
	}
	return type_builder.InternalNewTypeBuilder(tb), nil
}
func WithTypeBuilder(tb *TypeBuilder) CallOptionFunc {
	return func(o *callOption) {
		o.typeBuilder = tb.InternalExport()
	}
}