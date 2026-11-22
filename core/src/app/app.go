package app
import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"github.com/99designs/gqlgen/graphql"
	"github.com/clidey/whodb/core/src"
	"github.com/clidey/whodb/core/src/analytics"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/router"
	"github.com/clidey/whodb/core/src/settings"
	"github.com/clidey/whodb/core/src/source"
	_ "github.com/clidey/whodb/core/src/source/adapters"
)
const defaultPort = "8080"
func applicationVersion() string {
	if env.ApplicationVersion == "" {
		return "dev"
	}
	return env.ApplicationVersion
}
func runtimeMode() string {
	if env.GetIsDesktopMode() {
		return "desktop"
	}
	if env.GetIsCLIMode() {
		return "cli"
	}
	return "server"
}
func registeredPluginNames() []string {
	pluginNames := make([]string, 0, len(engine.RegisteredPlugins()))
	for _, p := range engine.RegisteredPlugins() {
		pluginNames = append(pluginNames, string(p.Type))
	}
	return pluginNames
}
func resolvePort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return defaultPort
	}
	return port
}
func welcomeBannerLines(port string) []string {
	title := "🎉 Welcome to WhoDB! 🎉"
	if env.IsEnterpriseEdition {
		title = "🎉 Welcome to WhoDB Enterprise! 🎉"
	}
	return []string{
		title,
		"Get started by visiting:",
		fmt.Sprintf("http://0.0.0.0:%s", port),
		"Explore and enjoy working with your databases!",
	}
}
func PopulateActiveDatabases() {
	for _, p := range engine.RegisteredPlugins() {
		env.ActiveDatabases = append(env.ActiveDatabases, string(p.Type))
	}
}
type AppConfig struct {
	Schema graphql.ExecutableSchema
	HTTPHandlers map[string]http.Handler
	Middlewares []func(http.Handler) http.Handler
	PublicPaths []string
	IsSetupMode bool
}
func Run(config AppConfig, staticFiles embed.FS) {
	showVersion := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(applicationVersion())
		return
	}
	PopulateActiveDatabases()
	defer log.CloseLogFile()
	appVersion := applicationVersion()
	log.Alwaysf("Starting WhoDB %s (log level: %s, set WHODB_LOG_LEVEL=warn or WHODB_LOG_LEVEL=error for quieter output)", appVersion, log.GetLevel())
	mode := runtimeMode()
	pluginNames := registeredPluginNames()
	log.Alwaysf("go=%s os=%s arch=%s mode=%s plugins=[%s]",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, mode, strings.Join(pluginNames, ", "))
	settingsCfg := settings.Get()
	if err := analytics.Initialize(analytics.Config{
		APIKey:      env.PosthogAPIKey,
		Host:        env.PosthogHost,
		Environment: env.ApplicationEnvironment,
		AppVersion:  env.ApplicationVersion,
	}); err != nil {
	} else {
		defer analytics.Shutdown()
	}
	analytics.SetEnabled(settingsCfg.MetricsEnabled)
	src.InitializeEngine()
	if err := settings.LoadProvidersFromFile(); err != nil {
		log.Warnf("Failed to load persisted AWS providers: %v", err)
	}
	if err := settings.InitAWSProvidersFromEnv(); err != nil {
		log.Warnf("Failed to initialize AWS providers from environment: %v", err)
	}
	if err := settings.LoadAzureProvidersFromFile(); err != nil {
		log.Warnf("Failed to load persisted Azure providers: %v", err)
	}
	if err := settings.InitAzureProvidersFromEnv(); err != nil {
		log.Warnf("Failed to initialize Azure providers from environment: %v", err)
	}
	if err := settings.LoadGCPProvidersFromFile(); err != nil {
		log.Warnf("Failed to load persisted GCP providers: %v", err)
	}
	if err := settings.InitGCPProvidersFromEnv(); err != nil {
		log.Warnf("Failed to initialize GCP providers from environment: %v", err)
	}
	r := router.InitializeRouter(config.Schema, config.HTTPHandlers, config.Middlewares, config.PublicPaths, staticFiles)
	port := resolvePort()
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      1 * time.Minute,
		IdleTimeout:       30 * time.Second,
	}
	go func() {
		log.Info("Almost there...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %s\n", err)
		}
	}()
	time.Sleep(2 * time.Second)
	for _, line := range welcomeBannerLines(port) {
		log.Always(line)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := srv.Shutdown(ctx); err != nil {
			log.Errorf("HTTP server shutdown error: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := source.Shutdown(ctx); err != nil {
			log.Errorf("source shutdown error: %v", err)
		}
	}()
	wg.Wait()
	close(quit)
	log.Info("Server exiting")
}