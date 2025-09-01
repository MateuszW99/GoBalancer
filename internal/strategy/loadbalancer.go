package strategy

import (
	"errors"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LoadBalancerStrategy interface {
	GetNextServer() (*server.Server, error)
	Done(*server.Server)
}

type LoadBalancer struct {
	strategy LoadBalancerStrategy
	logger   *zap.SugaredLogger
}

func SelectLoadBalancerWithStrategy(strat StrategyType, serverPool *server.ServerPool, logger *zap.SugaredLogger) (*LoadBalancer, error) {
	var strategy LoadBalancerStrategy

	switch strat {
	case RoundRobinStrategy:
		strategy = NewRoundRobinLoadBalancer(serverPool)
	case LeastConnections:
		strategy = NewLeastConnectionLoadBalancer(serverPool)
	default:
		return nil, errors.New("unknown strategy")
	}

	return &LoadBalancer{
		strategy: strategy,
		logger:   logger,
	}, nil
}

func (lb *LoadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	nextServer, err := lb.strategy.GetNextServer()

	if err != nil {
		lb.logger.Infow("failed to select next server", "error", err)
		http.Error(w, "internal nextServer error", http.StatusInternalServerError)
		return
	}
	lb.logger.Debugw("Calling next server", "serverName", nextServer.Name)

	defer lb.strategy.Done(nextServer)

	targetURL, err := url.Parse(nextServer.Url)
	if err != nil {
		lb.logger.Errorw("failed to parse nextServer url", "serverName", nextServer.Name)
		http.Error(w, "invalid backend URL", http.StatusInternalServerError)
		return
	}

	targetPath := strings.TrimRight(targetURL.String(), "/") + r.URL.Path
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetPath, r.Body)
	if err != nil {
		lb.logger.Errorw("failed to create request to backend", "serverName", nextServer.Name)
		http.Error(w, "failed to create request to backend", http.StatusInternalServerError)
		return
	}

	for k, values := range r.Header {
		for _, value := range values {
			req.Header.Add(k, value)
		}
	}

	req.Header.Set("X-Forwarded-For", r.RemoteAddr)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	res, err := client.Do(req)
	if err != nil {
		lb.logger.Errorw("failed to reach server", "serverName", nextServer.Name)
		http.Error(w, "failed to reach backend server", http.StatusBadGateway)
		return
	}

	byteResp, err := io.ReadAll(res.Body)
	if err != nil {
		lb.logger.Errorw("failed to read response from server", "serverName", nextServer.Name)
		http.Error(w, "failed to read response body", http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	for k, values := range res.Header {
		for _, value := range values {
			w.Header().Add(k, value)
		}
	}

	w.WriteHeader(res.StatusCode)
	fmt.Fprint(w, string(byteResp))
}
