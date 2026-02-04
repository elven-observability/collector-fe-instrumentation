// Main entrypoint for the Faro collector (frontend instrumentation → Loki).
package main

import (
	"log/slog"
	"os"

	httpadapter "collector-fe-instrumentation/internal/adapter/http"
	"collector-fe-instrumentation/internal/adapter/loki"
	"collector-fe-instrumentation/internal/config"
	"collector-fe-instrumentation/internal/usecase"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	lokiClient := loki.NewClient(cfg.LokiURL, cfg.LokiToken, cfg.LokiTimeout)
	collectorSvc := usecase.NewCollectorService(lokiClient, log)
	router := httpadapter.Router(cfg, collectorSvc, httpadapter.WithLogger(log))

	addr := ":" + cfg.HTTPPort
	slog.Info("listening", "addr", addr)
	if err := router.Run(addr); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
