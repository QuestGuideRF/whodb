package common
import (
	_ "github.com/clidey/whodb/core/src/bamlinit"
	"embed"
	"os"
	"strings"
	"github.com/99designs/gqlgen/graphql"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/clidey/whodb/core/src/analytics"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/router"
	"github.com/clidey/whodb/core/src/settings"
)
type RunConfig struct {
	Edition          string
	Title            string
	Assets           embed.FS
	Schema           graphql.ExecutableSchema
	InitializeEngine func()
}
func RunApp(config RunConfig) error {
	defer log.CloseLogFile()
	os.Setenv("WHODB_DESKTOP", "true")
	os.Setenv("WHODB_DISABLE_UPDATE_CHECK", "true")
	settingsCfg := settings.Get()
	if err := analytics.Initialize(analytics.Config{
		APIKey:      env.PosthogAPIKey,
		Host:        env.PosthogHost,
		Environment: env.ApplicationEnvironment,
		AppVersion:  env.ApplicationVersion,
	}); err != nil {
		log.WithError(err).Warn("Analytics: PostHog initialization failed, metrics disabled")
	} else {
		defer analytics.Shutdown()
	}
	analytics.SetEnabled(settingsCfg.MetricsEnabled)
	config.InitializeEngine()
	log.Infof("Running WhoDB Desktop %s Edition", strings.ToUpper(config.Edition))
	r := router.InitializeRouter(config.Schema, nil, nil, nil, config.Assets)
	app := NewApp(config.Edition)
	err := wails.Run(&options.App{
		Title:     config.Title,
		Width:     1400,
		Height:    900,
		MinWidth:  1024,
		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets:  config.Assets,
			Handler: r,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		OnBeforeClose:    app.beforeClose,
		OnDomReady:       app.DomReady,
		Bind: []any{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarDefault(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
	return err
}