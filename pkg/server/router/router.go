package router

import (
	"net/http"

	"github.com/rancher/apiserver/pkg/urlbuilder"
)

type RouterFunc func(h Handlers) http.Handler

type Handlers struct {
	K8sResource http.Handler
	APIRoot     http.Handler
	K8sProxy    http.Handler
	Next        http.Handler
	// ExtensionAPIServer serves under /ext. If nil, the default unknown path
	// handler is served.
	ExtensionAPIServer http.Handler
}

func Routes(h Handlers) http.Handler {
	m := http.NewServeMux()

	// Root handler with Accept header check for JSON
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			h.APIRoot.ServeHTTP(w, r)
			return
		}
		// Fallback to next for unmatched paths
		if h.Next != nil {
			h.Next.ServeHTTP(w, r)
		}
	})

	m.Handle("/v1", h.APIRoot)

	if h.ExtensionAPIServer != nil {
		m.Handle("/ext", http.StripPrefix("/ext", h.ExtensionAPIServer))
		m.Handle("/ext/", http.StripPrefix("/ext", h.ExtensionAPIServer))
	}

	// K8s resource routes
	m.Handle("/v1/{type}", h.K8sResource)
	m.Handle("/v1/{type}/{nameorns}", h.K8sResource)
	m.Handle("/v1/{type}/{namespace}/{name}", h.K8sResource)
	m.Handle("/v1/{type}/{namespace}/{name}/{link}", h.K8sResource)

	// K8s proxy routes
	m.Handle("/api", h.K8sProxy)
	m.Handle("/api/", h.K8sProxy)
	m.Handle("/apis/", h.K8sProxy)
	m.Handle("/openapi/", h.K8sProxy)
	m.Handle("/version/", h.K8sProxy)

	// Wrap with URL redirect rewrite middleware
	return urlbuilder.RedirectRewrite(m)
}
