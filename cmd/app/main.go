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
	"github.com/lmittmann/tint"
)

func main() {
	// handle shutdown signals
	shutdown := make(chan os.Signal, 3)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// load configuration
	c := internal.NewConfig()

	// set output logging
	level := slog.LevelDebug
	if c.Svc.Env != "dev" {
		level = slog.LevelInfo
	}

	slog.SetDefault(
		slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{Level: level}),
		),
	)

	noopDependencies := handler.Dependencies{}
	// set up the route handler and server
	app := handler.New(noopDependencies)
	srv := &http.Server{
		Addr:         c.Svc.Address,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      app,
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
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Print("Server stopped successfully")
}
