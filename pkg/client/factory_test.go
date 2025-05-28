package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rancher/apiserver/pkg/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func TestWithQPSAndBurst(t *testing.T) {
	fo := defaultFactoryOptions()

	assert.Equal(t, defaultQPS, fo.qps)
	assert.Equal(t, defaultBurst, fo.burst)

	WithQPSAndBurst(50, 20, true)(fo)

	assert.Equal(t, float32(50.0), fo.qps)
	assert.Equal(t, 20, fo.burst)
	assert.True(t, fo.sharedQPS)
}

func TestNewFactoryWithOptions(t *testing.T) {
	cfg := &rest.Config{
		QPS:   5.0,
		Burst: 10,
	}
	f, err := NewFactory(cfg, false, WithQPSAndBurst(50, 20, false))
	assert.NoError(t, err)

	assert.Equal(t, float32(50.0), f.clientCfg.QPS)
	assert.Equal(t, 20, f.clientCfg.Burst)
}

func TestFactoryBalancing(t *testing.T) {
	queryAPI := func(t *testing.T, wg *sync.WaitGroup, f *Factory, failErrors bool) {
		defer wg.Done()
		dc, err := f.DynamicClient(&types.APIRequest{}, nil)
		assert.NoError(t, err)

		for i := 0; i < 20; i++ {
			_, err = dc.Resource(schema.GroupVersionResource{
				Resource: "namespaces",
				Version:  "v1",
			}).Get(context.TODO(), "testing", metav1.GetOptions{})
			if err != nil && failErrors {
				t.Error(err)
			}
		}
	}

	t.Run("Without distributing the QPS across clients", func(t *testing.T) {
		var errorCount int32
		srv := startTestServer(t, rate.Limit(10), &errorCount)
		cfg := &rest.Config{Host: srv.URL}

		f, err := NewFactory(cfg, false, WithQPSAndBurst(9.0, 1, false))
		assert.NoError(t, err)

		wg := &sync.WaitGroup{}
		concurrency := 5
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go queryAPI(t, wg, f, false)
		}
		wg.Wait()

		assert.NotZero(t, errorCount)
	})

	t.Run("When distributing the QPS across clients", func(t *testing.T) {
		var errorCount int32
		srv := startTestServer(t, rate.Limit(10), &errorCount)
		cfg := &rest.Config{Host: srv.URL}
		f, err := NewFactory(cfg, false, WithQPSAndBurst(9.0, 1, true))
		assert.NoError(t, err)

		wg := &sync.WaitGroup{}
		concurrency := 5
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go queryAPI(t, wg, f, true)
		}
		wg.Wait()

		assert.Equal(t, int32(0), errorCount)
	})

	t.Run("When the Distributed QPS is above the rate limit", func(t *testing.T) {
		var errorCount int32
		srv := startTestServer(t, rate.Limit(10), &errorCount)
		cfg := &rest.Config{Host: srv.URL}

		f, err := NewFactory(cfg, false, WithQPSAndBurst(11.0, 1, true))
		assert.NoError(t, err)

		wg := &sync.WaitGroup{}
		concurrency := 5
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go queryAPI(t, wg, f, false)
		}
		wg.Wait()

		assert.NotZero(t, errorCount)
	})
}

// This provides a fake K8s API server that uses the provided rate.Limit to
// rate-limit requests, responding with 429 if the rate-limiter is limiting
// requests.
//
// It only allows getting a "testing" namespace and responds with a hard-coded
// Namespace resource in JSON format.
//
// The errors value passed in will be incremented every time a 429 response is
// returned to the client (client-go will consume some 429 responses).
func startTestServer(t *testing.T, limit rate.Limit, errors *int32) *httptest.Server {
	start := time.Now()
	rl := rate.NewLimiter(limit, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/namespaces/testing" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if !rl.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			atomic.AddInt32(errors, 1)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"apiVersion":"v1","kind":"Namespace","metadata":{"creationTimestamp":"2025-03-27T10:02:44Z","labels":{"kubernetes.io/metadata.name":"testing"},"name":"testing","resourceVersion":"3319","uid":"5b59c95b-6a85-4107-a57e-2240e46086e8"},"spec":{"finalizers":["kubernetes"]},"status": {"phase": "Active"}}`))
	}))
	t.Cleanup(func() {
		t.Logf("%v errors in %v", *errors, time.Since(start))
		ts.Close()
	})

	return ts
}
