package main
import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
	"errors"
	"github.com/clidey/whodb/core/graph"
	"github.com/clidey/whodb/core/src"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/router"
)
var staticFilesTest embed.FS
const defaultPortTest = "8080"
var srv *http.Server
func TestMain(m *testing.M) {
	log.Info("Starting WhoDB in test mode (Ctrl+C to exit)...")
	src.InitializeEngine()
	schema := graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}})
	r := router.InitializeRouter(schema, nil, nil, nil, staticFilesTest)
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPortTest
	}
	srv = &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      1 * time.Minute,
		IdleTimeout:       30 * time.Second,
	}
	go func() {
		log.Info("Server starting...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s", err)
			os.Exit(1)
		}
	}()
	log.Infof("🎉 WhoDB test server running at http://localhost:%s 🎉", port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Received shutdown signal (Ctrl+C). Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Graceful shutdown failed: %v", err)
	}
	log.Info("Test server shut down. Exiting and writing coverage.")
	os.Exit(m.Run())
}