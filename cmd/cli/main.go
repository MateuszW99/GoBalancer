package main

import (
	"context"
	"errors"
	"github.com/MateuszW99/GoBalancer/internal/app"
	"github.com/MateuszW99/GoBalancer/internal/config"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()

	appCfg, err := config.Load(sugar)
	if err != nil {
		sugar.Fatalw("failed to build app config", "err", err)
	}

	application, err := app.NewApp(
		app.WithConfig(appCfg),
		app.WithLogger(sugar),
	)
	if err != nil {
		sugar.Fatalw("failed to build app", "err", err)
	}

	if err := application.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		sugar.Fatalw("app run failed", "err", err)
	}
}
