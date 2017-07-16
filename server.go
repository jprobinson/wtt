package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/jprobinson/gosubway"
)

func init() {
	key := os.Getenv("MTA_KEY")
	r := mux.NewRouter()
	r.HandleFunc("http://wtt.nyc/.well-known/acme-challenge/ZPU1wVSp4PYGc7EwvgfQ2sDzZzDTLZJU4rn6TeN8WBk",
		func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "ZPU1wVSp4PYGc7EwvgfQ2sDzZzDTLZJU4rn6TeN8WBk.sz06j6KUG_9nJGNedr8vGy4DZ6mo_af0MFcEBuWk0Uo")
		})
	r.HandleFunc("/.well-known/acme-challenge/SJhmh-8U1pNgTtERQ3I7vCO78sv0n21pd5dVntQJ2K0",
		func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "SJhmh-8U1pNgTtERQ3I7vCO78sv0n21pd5dVntQJ2K0.sz06j6KUG_9nJGNedr8vGy4DZ6mo_af0MFcEBuWk0Uo")
		})
	r.HandleFunc("/svc/subway-api/v1/next-trains/{line}/{stopID}", nextTrains(key)).Methods("GET")
	r.HandleFunc("/_ah/warmup",
		func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
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
		line := strings.ToUpper(vars["line"])

		ctx := appengine.NewContext(r)
		feed, err := getFeed(ctx, key, (line == ltrain))
		if err != nil {
			log.Errorf(ctx, "unable to get subway feed: %s", err)
			http.Error(w, "unable to read subway feed", http.StatusInternalServerError)
			return
		}

		alerts, north, south := feed.NextTrainTimes(stop, line)
		resp := nextTrainResp{
			Northbound: north,
			Southbound: south,
			Alerts:     alerts,
		}
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Errorf(ctx, "unable to encode response: ", err)
		}
	}
}

const (
	timeout     = 1 * time.Second
	maxAttempts = 10
	backoffStep = 50 * time.Millisecond
)

// retries until it hits max attempts or a context timeout
func getFeed(ctx context.Context, key string, l bool) (*gosubway.FeedMessage, error) {
	var (
		feed *gosubway.FeedMessage
		err  error
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// retry backoff
		time.Sleep(time.Duration((attempt - 1)) * backoffStep)
		// attempt to get feed
		feed, err = gosubway.GetFeed(ctx, key, l)
		if err == nil ||
			(err != nil && strings.Contains(err.Error(), "deadline exceeded")) {
			break
		}
		log.Errorf(ctx, "unable to get mta feed on attempt %d: %s", attempt, err)
	}
	return feed, err
}

type nextTrainResp struct {
	Northbound []time.Time       `json:"northbound"`
	Southbound []time.Time       `json:"southbound"`
	Alerts     []*gosubway.Alert `json:"alerts"`
}
