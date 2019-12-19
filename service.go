package wtt

import (
	"context"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/NYTimes/gizmo/server/kit"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc"
)

type service struct {
	key  string
	base string

	hc *http.Client
}

func NewService() kit.Service {
	return &service{
		key:  os.Getenv("MTA_KEY"),
		base: os.Getenv("BASE_PATH"),
		hc:   &http.Client{Timeout: 2 * time.Second},
	}
}

func (s service) HTTPOptions() []httptransport.ServerOption {
	return []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(func(ctx context.Context, err error, w http.ResponseWriter) {
			httptransport.EncodeJSONResponse(ctx, w, err)
		}),
	}
}

func (s *service) HTTPMiddleware(h http.Handler) http.Handler {
	return h
}

func (s *service) Middleware(e endpoint.Endpoint) endpoint.Endpoint {
	return e
}

func (s *service) HTTPRouterOptions() []kit.RouterOption {
	return nil
}

func (s service) HTTPEndpoints() map[string]map[string]kit.HTTPEndpoint {
	return map[string]map[string]kit.HTTPEndpoint{
		"/": {
			"GET": {
				Endpoint: static,
				Encoder:  serveFile(s.base+"/pages/index.html", htmlct),
			},
		},
		"/svc/subway-api/v1/next-trains/{line}/{stopID}": {
			"GET": {
				Endpoint: s.getNextTrains,
				Decoder:  decodeNextTrains,
			},
		},
		"/css/{name}": {
			"GET": {
				Endpoint: static,
				Encoder:  serveDir(s.base+"/css/", cssct),
			},
		},
		"/js/{name}": {
			"GET": {
				Endpoint: static,
				Encoder:  serveDir(s.base+"/js/", jsct),
			},
		},
		"/js/vendor/{name}": {
			"GET": {
				Endpoint: static,
				Encoder:  serveDir(s.base+"/js/vendor/", jsct),
			},
		},
		"/data/{name}": {
			"GET": {
				Endpoint: static,
				Encoder:  serveDir(s.base+"/data/", jsonct),
			},
		},
		"/images/{name}": {
			"GET": {
				Endpoint: static,
				Encoder:  serveDir(s.base+"/images/", ""),
			},
		},
		"/robots.txt": {
			"GET": {
				Endpoint: static,
				Encoder:  serveFile(s.base+"/pages/robots.txt", plainct),
			},
		},
		"/humans.txt": {
			"GET": {
				Endpoint: static,
				Encoder:  serveFile(s.base+"/pages/humans.txt", plainct),
			},
		},
		"/terms.html": {
			"GET": {
				Endpoint: static,
				Encoder:  serveFile(s.base+"/pages/terms.html", htmlct),
			},
		},
		"/privacy.html": {
			"GET": {
				Endpoint: static,
				Encoder:  serveFile(s.base+"/pages/privacy.html", htmlct),
			},
		},
	}
}

const (
	plainct = "text/plain"
	htmlct  = "text/html"
	jsct    = "text/javascript;charset=UTF-8"
	jsonct  = "application/json"
	cssct   = "text/css"
)

func static(ctx context.Context, r interface{}) (interface{}, error) {
	return r, nil
}

func serveDir(d string, ct string) httptransport.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, rq interface{}) error {
		if ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		r := rq.(*http.Request)
		http.ServeFile(w, r, d+path.Base(r.URL.Path))
		return nil
	}
}

func serveFile(f string, ct string) httptransport.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, r interface{}) error {
		if ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		http.ServeFile(w, r.(*http.Request), f)
		return nil
	}
}

func (s *service) RPCMiddleware() grpc.UnaryServerInterceptor {
	return nil
}

func (s *service) RPCServiceDesc() *grpc.ServiceDesc {
	return nil
}

func (s *service) RPCOptions() []grpc.ServerOption {
	return nil
}
