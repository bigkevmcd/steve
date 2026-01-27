package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	// Create mock handlers that track if they were called
	apiRootCalled := false
	k8sResourceCalled := false
	k8sProxyCalled := false
	nextCalled := false
	extCalled := false

	apiRootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiRootCalled = true
		w.WriteHeader(http.StatusOK)
	})

	k8sResourceHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k8sResourceCalled = true
		w.WriteHeader(http.StatusOK)
	})

	k8sProxyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k8sProxyCalled = true
		w.WriteHeader(http.StatusOK)
	})

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	extHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extCalled = true
		w.WriteHeader(http.StatusOK)
	})

	tests := map[string]struct {
		path          string
		handlers      Handlers
		expectAPIRoot bool
		expectK8s     bool
		expectProxy   bool
		expectNext    bool
		expectExt     bool
	}{
		"root path calls APIRoot": {
			path: "/",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectAPIRoot: true,
		},
		"get /v1 calls APIRoot": {
			path: "/v1",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectAPIRoot: true,
		},
		"get /v1/pods calls K8sResource": {
			path: "/v1/pods",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectK8s: true,
		},
		"get /v1/pods/default calls K8sResource": {
			path: "/v1/pods/default",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectK8s: true,
		},
		"get /v1/pods/default/mypod calls K8sResource": {
			path: "/v1/pods/default/mypod",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectK8s: true,
		},
		"get /v1/pods/default/mypod/logs calls K8sResource": {
			path: "/v1/pods/default/mypod/logs",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectK8s: true,
		},
		"get /api calls K8sProxy": {
			path: "/api",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectProxy: true,
		},
		"get /api/ calls K8sProxy": {
			path: "/api/v1",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectProxy: true,
		},
		"get /apis/ calls K8sProxy": {
			path: "/apis/apps/v1",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectProxy: true,
		},
		"get /openapi/ calls K8sProxy": {
			path: "/openapi/v2",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectProxy: true,
		},
		"get /version/ calls K8sProxy": {
			path: "/version/",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
			},
			expectProxy: true,
		},
		"unmatched path calls Next handler": {
			path: "/unknown/path",
			handlers: Handlers{
				APIRoot:     apiRootHandler,
				K8sResource: k8sResourceHandler,
				K8sProxy:    k8sProxyHandler,
				Next:        nextHandler,
			},
			expectNext: true,
		},
		"get /ext calls ExtensionAPIServer": {
			path: "/ext",
			handlers: Handlers{
				APIRoot:            apiRootHandler,
				K8sResource:        k8sResourceHandler,
				K8sProxy:           k8sProxyHandler,
				ExtensionAPIServer: extHandler,
			},
			expectExt: true,
		},
		"get /ext/ calls ExtensionAPIServer": {
			path: "/ext/some/path",
			handlers: Handlers{
				APIRoot:            apiRootHandler,
				K8sResource:        k8sResourceHandler,
				K8sProxy:           k8sProxyHandler,
				ExtensionAPIServer: extHandler,
			},
			expectExt: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Reset flags
			apiRootCalled = false
			k8sResourceCalled = false
			k8sProxyCalled = false
			nextCalled = false
			extCalled = false

			router := Routes(tt.handlers)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if apiRootCalled != tt.expectAPIRoot {
				t.Errorf("APIRoot called = %v, want %v", apiRootCalled, tt.expectAPIRoot)
			}
			if k8sResourceCalled != tt.expectK8s {
				t.Errorf("K8sResource called = %v, want %v", k8sResourceCalled, tt.expectK8s)
			}
			if k8sProxyCalled != tt.expectProxy {
				t.Errorf("K8sProxy called = %v, want %v", k8sProxyCalled, tt.expectProxy)
			}
			if nextCalled != tt.expectNext {
				t.Errorf("Next called = %v, want %v", nextCalled, tt.expectNext)
			}
			if extCalled != tt.expectExt {
				t.Errorf("ExtensionAPIServer called = %v, want %v", extCalled, tt.expectExt)
			}
		})
	}
}
