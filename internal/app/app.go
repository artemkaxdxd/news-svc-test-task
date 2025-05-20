package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	handlerpost "news-svc/internal/controller/web/v1/post"
	svcpost "news-svc/internal/service/post"
	repopost "news-svc/internal/storage/mongo/post"

	"news-svc/config"
	"news-svc/pkg/httpserver"
	"news-svc/pkg/mongo"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg config.Config) {
	logLevel := slog.LevelInfo
	if cfg.Server.IsDev {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, err := mongo.New(ctx,
		cfg.Mongo.User,
		cfg.Mongo.Password,
		cfg.Mongo.Host,
		cfg.Mongo.Port,
		cfg.Mongo.Name,
	)
	if err != nil {
		logger.Error("unable to connect to MongoDB", "err", err)
		return
	}

	logger.Info("connected to MongoDB", "db", cfg.Mongo.Name)

	defer func() {
		if err := client.Close(ctx); err != nil {
			logger.Error("error disconnecting MongoDB", "err", err)
		}
	}()

	postRepo := repopost.New(client.Instance())
	postSvc := svcpost.New(postRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	handlerpost.InitHandler(mux, postSvc, logger)

	srv := httpserver.New(
		mux,
		httpserver.Port(cfg.Server.Port),
	)

	logger.Info("starting http server", "port", cfg.Server.Port)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-srv.Notify():
		logger.Error("HTTP server error", "err", err)
	}

	if err := srv.Shutdown(); err != nil {
		logger.Error("server shutdown error", "err", err)
	} else {
		logger.Info("server stopped gracefully")
	}
}
