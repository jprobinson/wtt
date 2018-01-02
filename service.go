package main

import (
	"context"
	"net/http"
	"os"

	"github.com/NYTimes/marvin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type service struct {
	key string
}

func NewService() marvin.JSONService {
	return &service{key: os.Getenv("MTA_KEY")}
}

func (s service) Options() []httptransport.ServerOption {
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

func (s *service) RouterOptions() []marvin.RouterOption {
	return nil
}

func (s service) JSONEndpoints() map[string]map[string]marvin.HTTPEndpoint {
	return map[string]map[string]marvin.HTTPEndpoint{
		"/svc/subway-api/v1/next-trains/{line}/{stopID}": {
			"GET": {
				Endpoint: s.getNextTrains,
				Decoder:  decodeNextTrains,
			},
		},
		"/svc/subway-api/v1/dialogflow": {
			"POST": {
				Endpoint: s.postDialogflow,
				Decoder:  decodeDialogflow,
			},
		},
	}
}
