package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gerbenjacobs/gerben.dev/handler"
	"github.com/gerbenjacobs/gerben.dev/internal"
	"github.com/go-slog/otelslog"
	"github.com/lmittmann/tint"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	// handle shutdown signals
	shutdown := make(chan os.Signal, 3)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// load configuration
	c := internal.NewConfig()

	// set output logging
	logger := slog.New(otelslog.NewHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	if c.Svc.Env == "dev" {
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	// load opentelemetry
	otelShutdown, err := internal.SetupOTelSDK(context.Background(), c.Svc.Env, "gerben.dev", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	// create caches
	if err := internal.CreateCaches(); err != nil {
		log.Fatalf("failed to create cache: %v", err)
	}

	dependencies := handler.Dependencies{
		SecretKey: c.Svc.SecretToken,
	}

	// set up the route handler and server
	app := handler.New(c.Svc.Env, dependencies)
	appHandler := otelhttp.NewHandler(app, "server", otelhttp.WithSpanOptions(trace.WithSpanKind(trace.SpanKindServer)))
	srv := &http.Server{
		Addr:         c.Svc.Address,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      appHandler,
	}

	// start running the server
	go func() {
		log.Print("Server started on http://" + srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to listen: %v", err)
		}
	}()

	// wait for shutdown signals
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := otelShutdown(ctx); err != nil {
		slog.Error("Failed to shutdown OpenTelemetry", "error", err)
		os.Exit(1)
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Print("Server stopped successfully")
}
