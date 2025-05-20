package main

import (
	"log/slog"
	"news-svc/config"
	"news-svc/internal/app"
	"os"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		slog.Error("config fill", "err", err)
		os.Exit(1)
	}

	app.Run(cfg)
}
