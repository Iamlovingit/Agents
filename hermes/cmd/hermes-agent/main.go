package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"clawmanager-hermes-image/internal/agent"
)

var version = "dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := agent.LoadConfig(version)
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(2)
	}

	if !cfg.Enabled {
		logger.Info("ClawManager Hermes agent disabled")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runtimeAgent := agent.New(cfg, logger)
	if err := runtimeAgent.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("agent stopped", "error", err)
		os.Exit(1)
	}
}
