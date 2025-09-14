package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/api"
	"github.com/MateuszW99/GoBalancer/internal/config"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

type App struct {
	Logger      *zap.SugaredLogger
	Cfg         *config.AppConfig
	LbProxy     *http.Server
	AdminServer *http.Server
}

func NewApp(cfg *config.AppConfig, logger *zap.SugaredLogger) (*App, error) {
	pool := selectServerPool(cfg.ServerPools)
	lbStrat, err := strategy.SelectLoadBalancerWithStrategy(strategy.ParseStrategyType(pool.Strategy), pool, logger)
	if err != nil {
		logger.Fatalw("failed to select strategy", "error", err)
	}

	lbServer := configureLoadBalancerServer(cfg.LbPort, lbStrat)
	adminServer := configureAdminServer(cfg, logger)

	return &App{
		Logger:      logger,
		Cfg:         cfg,
		LbProxy:     lbServer,
		AdminServer: adminServer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		a.Logger.Infow("starting load balancer on port", "port", a.Cfg.LbPort)
		err := a.LbProxy.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	group.Go(func() error {
		a.Logger.Infow("starting admin api on port", "port", a.Cfg.AdminPort)
		err := a.AdminServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	group.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.LbProxy.Shutdown(shutdownCtx)
		_ = a.AdminServer.Shutdown(shutdownCtx)
		return nil
	})

	return group.Wait()
}

func selectServerPool(pools []*server.ServerPool) *server.ServerPool {
	return pools[0]
}

func configureLoadBalancerServer(port int, lbStrat *strategy.LoadBalancer) *http.Server {
	router := api.NewLoadBalancerRouter(lbStrat)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	return srv
}

func configureAdminServer(cfg *config.AppConfig, logger *zap.SugaredLogger) *http.Server {
	adminApi := api.NewAdminApi(cfg.ServerPools, logger)
	router := api.NewAdminRouter(adminApi)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.AdminPort),
		Handler: router,
	}
	return srv
}
