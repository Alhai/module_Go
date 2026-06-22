package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/alhai/urlwatch/internal/api"
	"github.com/alhai/urlwatch/internal/checker"
	"github.com/alhai/urlwatch/internal/store"
)

func main() {
	logger := buildLogger()
	slog.SetDefault(logger)

	chkr := checker.NewHTTPChecker()
	st := store.NewMemoryStore()
	h := api.NewHandler(chkr, st, logger)
	router := api.NewRouter(h, logger)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	logger.Info("starting urlwatch", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func buildLogger() *slog.Logger {
	level := slog.LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
