package strategy

import (
	"errors"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LoadBalancerStrategy interface {
	GetNextServer() (*server.Server, error)
}

type LoadBalancer struct {
	strategy LoadBalancerStrategy
}

func SelectLoadBalancerWithStrategy(strat StrategyType, serverPool *server.ServerPool) (*LoadBalancer, error) {
	switch strat {
	case RoundRobinStrategy:
		return &LoadBalancer{
			strategy: NewRoundRobinLoadBalancer(serverPool),
		}, nil
	default:
		return nil, errors.New("unknown strategy")
	}
}

func (lb *LoadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	nextServer, err := lb.strategy.GetNextServer()

	if err != nil {
		http.Error(w, "internal nextServer error", http.StatusInternalServerError)
		return
	}
	log.Printf("Calling next server %v", nextServer.Name)

	targetURL, err := url.Parse(nextServer.Url)
	if err != nil {
		http.Error(w, "invalid backend URL", http.StatusInternalServerError)
		return
	}

	targetPath := strings.TrimRight(targetURL.String(), "/") + r.URL.Path
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetPath, r.Body)
	if err != nil {
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
		http.Error(w, "failed to reach backend nextServer", http.StatusBadGateway)
		return
	}

	byteResp, err := io.ReadAll(res.Body)
	if err != nil {
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
