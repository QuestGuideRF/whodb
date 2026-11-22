package router
import (
	"embed"
	"net/http"
	"time"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/clidey/whodb/core/graph"
	"github.com/clidey/whodb/core/src/auth"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)
type OAuthLoginUrl struct {
	Url string `json:"url"`
}
func NewGraphQLServer(es graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(es)
	srv.AddTransport(&transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(&transport.Options{})
	srv.AddTransport(&transport.GET{})
	srv.AddTransport(&transport.POST{})
	srv.AddTransport(&transport.MultipartForm{
		MaxUploadSize: 250 * 1024 * 1024,
	})
	srv.Use(extension.FixedComplexityLimit(100))
	if env.IsDevelopment {
		srv.Use(extension.Introspection{})
	}
	return srv
}
func setupServer(router *chi.Mux, schema graphql.ExecutableSchema, httpHandlers map[string]http.Handler, staticFiles embed.FS) {
	if !env.IsAPIGatewayEnabled {
		fileServer(router, staticFiles)
	}
	server := NewGraphQLServer(schema)
	server.AddTransport(&transport.Websocket{})
	graph.SetupHTTPServer(router)
	setupPlaygroundHandler(router, server)
	for path, h := range httpHandlers {
		router.Handle(path, h)
	}
}
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}
func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
func (w *statusResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
func healthCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.LogAccess(r.Method, r.URL.Path, sw.statusCode, time.Since(start), r.Host, r.RemoteAddr)
	})
}
func setupMiddlewares(router *chi.Mux, additionalMiddlewares []func(http.Handler) http.Handler, publicPaths []string) {
	allowedOrigins := env.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, "https://*", "http://*")
	}
	middlewares := []func(http.Handler) http.Handler{
		accessLogMiddleware,
		healthCheckMiddleware,
		middleware.ThrottleBacklog(100, 50, time.Second*5),
		middleware.RequestID,
		middleware.RealIP,
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.Timeout(90 * time.Second),
		cors.Handler(cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{},
			AllowCredentials: true,
			MaxAge:           300,
		}),
		contextMiddleware,
	}
	middlewares = append(middlewares, additionalMiddlewares...)
	if len(publicPaths) > 0 {
		bypassSet := make(map[string]struct{}, len(publicPaths))
		for _, p := range publicPaths {
			bypassSet[p] = struct{}{}
		}
		middlewares = append(middlewares, func(next http.Handler) http.Handler {
			authNext := auth.AuthMiddleware(next)
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, ok := bypassSet[r.URL.Path]; ok {
					next.ServeHTTP(w, r)
					return
				}
				authNext.ServeHTTP(w, r)
			})
		})
	} else {
		middlewares = append(middlewares, auth.AuthMiddleware)
	}
	router.Use(middlewares...)
}
func wrapWithBasePath(handler http.Handler, basePath string) *chi.Mux {
	router := chi.NewRouter()
	redirectToBasePath := func(w http.ResponseWriter, r *http.Request) {
		target := basePath + "/"
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	}
	router.Get(basePath, redirectToBasePath)
	router.Head(basePath, redirectToBasePath)
	router.Handle("/health", handler)
	router.Handle(basePath+"/*", http.StripPrefix(basePath, handler))
	return router
}
func InitializeRouter(schema graphql.ExecutableSchema, httpHandlers map[string]http.Handler, additionalMiddlewares []func(http.Handler) http.Handler, publicPaths []string, staticFiles embed.FS) *chi.Mux {
	router := chi.NewRouter()
	setupMiddlewares(router, additionalMiddlewares, publicPaths)
	setupServer(router, schema, httpHandlers, staticFiles)
	if env.BasePath == "" {
		return router
	}
	if env.IsAPIGatewayEnabled || !hasEmbeddedFrontend(staticFiles) {
		log.Warnf("Ignoring WHODB_BASE_PATH=%s because bundled frontend assets are not being served", env.BasePath)
		return router
	}
	return wrapWithBasePath(router, env.BasePath)
}