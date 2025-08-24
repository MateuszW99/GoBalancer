package integration

import (
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/gavv/httpexpect"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_RoundRobin(t *testing.T) {
	backendSrv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello from backend server 1"))
		if err != nil {
			return
		}
	}))
	defer backendSrv1.Close()

	backendSrv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello from backend server 2"))
		if err != nil {
			return
		}
	}))
	defer backendSrv2.Close()

	s1 := &server.Server{Url: backendSrv1.URL, HealthcheckUrl: "", IsHealthy: true}
	s2 := &server.Server{Url: backendSrv2.URL, HealthcheckUrl: "", IsHealthy: true}
	lb := startLoadBalancer(t, []*server.Server{s1, s2}, "rb")

	e := httpexpect.New(t, lb.URL)

	e.GET("/").Expect().Status(http.StatusOK).Body().Contains("hello from backend server 1")
	e.GET("/").Expect().Status(http.StatusOK).Body().Contains("hello from backend server 2")
	e.GET("/").Expect().Status(http.StatusOK).Body().Contains("hello from backend server 1")
}

func Test_RoundRobinHealthFlip(t *testing.T) {
	healthySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello from healthy server"))
		if err != nil {
			return
		}
	}))
	defer healthySrv.Close()

	unhealthySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer unhealthySrv.Close()

	s1 := &server.Server{Url: healthySrv.URL, HealthcheckUrl: "", IsHealthy: true, Name: "healthy"}
	s2 := &server.Server{Url: unhealthySrv.URL, HealthcheckUrl: "", IsHealthy: true, Name: "unhealthy"}
	lb := startLoadBalancer(t, []*server.Server{s1, s2}, "rb")

	waitUntilServerUnhealthy(t, s2, e2eTestCfg)

	e := httpexpect.New(t, lb.URL)
	for i := 0; i < 3; i++ {
		e.GET("/").Expect().Status(http.StatusOK).Body().Contains("healthy")
	}
}
