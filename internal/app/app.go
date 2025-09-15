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

type Option func(*options) error

type options struct {
	cfg                *config.AppConfig
	logger             *zap.SugaredLogger
	lbServer           *http.Server
	adminServer        *http.Server
	strategySelector   func(pool *server.ServerPool, logger *zap.SugaredLogger) (*strategy.LoadBalancer, error)
	lbServerFactory    func(port int, lb *strategy.LoadBalancer) *http.Server
	adminServerFactory func(port int, pools []*server.ServerPool, logger *zap.SugaredLogger) *http.Server
}

func defaultOptions() *options {
	return &options{
		strategySelector: func(pool *server.ServerPool, logger *zap.SugaredLogger) (*strategy.LoadBalancer, error) {
			return strategy.SelectLoadBalancerWithStrategy(strategy.ParseStrategyType(pool.Strategy), pool, logger)
		},
		lbServerFactory: func(port int, lb *strategy.LoadBalancer) *http.Server {
			return configureLoadBalancerServer(port, lb)
		},
		adminServerFactory: func(port int, pools []*server.ServerPool, logger *zap.SugaredLogger) *http.Server {
			return configureAdminServer(port, pools, logger)
		},
	}
}

func WithConfig(cfg *config.AppConfig) Option {
	return func(o *options) error {
		o.cfg = cfg
		return nil
	}
}

func WithLogger(logger *zap.SugaredLogger) Option {
	return func(o *options) error {
		o.logger = logger
		return nil
	}
}

func WithStrategySelector(f func(*server.ServerPool, *zap.SugaredLogger) (*strategy.LoadBalancer, error)) Option {
	return func(o *options) error {
		o.strategySelector = f
		return nil
	}
}

func WithLoadBalancerFactory(f func(int, *strategy.LoadBalancer) *http.Server) Option {
	return func(o *options) error {
		o.lbServerFactory = f
		return nil
	}
}
func WithAdminServerFactory(f func(int, []*server.ServerPool, *zap.SugaredLogger) *http.Server) Option {
	return func(o *options) error {
		o.adminServerFactory = f
		return nil
	}
}

func NewApp(opts ...Option) (*App, error) {
	o := defaultOptions()
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	if o.lbServer == nil {
		pool := selectServerPool(o.cfg.ServerPools)
		lb, err := o.strategySelector(pool, o.logger)
		if err != nil {
			return nil, fmt.Errorf("select strategy: %w", err)
		}
		o.lbServer = o.lbServerFactory(o.cfg.LbPort, lb)
	}

	if o.adminServer == nil {
		o.adminServer = o.adminServerFactory(o.cfg.AdminPort, o.cfg.ServerPools, o.logger)
	}

	return &App{
		Logger:      o.logger,
		Cfg:         o.cfg,
		LbProxy:     o.lbServer,
		AdminServer: o.adminServer,
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

func configureAdminServer(port int, serverPool []*server.ServerPool, logger *zap.SugaredLogger) *http.Server {
	adminApi := api.NewAdminApi(serverPool, logger)
	router := api.NewAdminRouter(adminApi)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	return srv
}
