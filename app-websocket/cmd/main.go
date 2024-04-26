package main

import (
	"app-websocket/internal/components"
	"app-websocket/internal/config"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	logger := components.SetupLogger(cfg.Env)

	components, err := components.InitComponents(cfg, logger)
	if err != nil {
		logger.Error("bad configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer components.Shutdown()

	eg, ctx := errgroup.WithContext(context.Background())
	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)

	eg.Go(func() error {
		return components.HttpServer.Run(ctx)
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case s := <-sigQuit:
			logger.Info("Captured signal", slog.String("signal", s.String()))
			return fmt.Errorf("captured signal: %v", s)
		}
	})

	err = eg.Wait()
	logger.Info("Gracefully shutting down the servers", slog.String("error", err.Error()))
}
