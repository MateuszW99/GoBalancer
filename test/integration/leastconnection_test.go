package integration

import (
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func Test_LeastConnection(t *testing.T) {
	backendSrv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
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
	lb := startLoadBalancer(t, []*server.Server{s1, s2}, "lc")

	client := &http.Client{Timeout: 1 * time.Second}
	const requestNum = 20
	var wg sync.WaitGroup
	wg.Add(requestNum)
	results := make(chan string, requestNum)

	for i := 0; i < requestNum; i++ {
		go func() {
			defer wg.Done()
			resp, err := client.Get(lb.URL + "/")
			if err != nil {
				results <- fmt.Sprintf("Err: %v", err.Error())
				return
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					panic(err)
				}
			}(resp.Body)
			body, _ := io.ReadAll(resp.Body)
			results <- string(body)
		}()
	}

	wg.Wait()
	close(results)

	slow := 0
	fast := 0
	errors := 0

	for resp := range results {
		if strings.HasPrefix(resp, "ERR") {
			errors++
			continue
		}
		if resp == "hello from backend server 1" {
			slow++
			continue
		}
		if resp == "hello from backend server 2" {
			fast++
			continue
		}
	}

	if errors > 0 {
		t.Fatalf("got %d request errors", errors)
	}

	require.Less(t, slow, fast)
}
