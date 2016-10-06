package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/gorilla/mux"
	"github.com/jprobinson/gosubway"
)

func init() {
	key := os.Getenv("MTA_KEY")
	r := mux.NewRouter()
	r.HandleFunc("/svc/subway-api/v1/next-trains/{feed}/{stopID}", nextTrains(key)).Methods("GET")
	http.Handle("/", r)
}

const (
	ltrain = "L"
	other  = "123456S"
)

func nextTrains(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "application/json; charset=utf8")

		vars := mux.Vars(r)
		stop := vars["stopID"]
		feedType := vars["feed"]

		ctx := appengine.NewContext(r)
		feed, err := gosubway.GetFeed(ctx, key, (feedType == ltrain))
		if err != nil {
			log.Errorf(ctx, "unable to get subway feed: ", err)
			http.Error(w, "dammit", http.StatusBadRequest)
			return
		}

		north, south := feed.NextTrainTimes(stop)
		resp := nextTrainResp{north, south}
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Errorf(ctx, "unable to encode response: ", err)
		}

	}
}

type nextTrainResp struct {
	Northbound []time.Time `json:"northbound"`
	Southbound []time.Time `json:"southbound"`
}
