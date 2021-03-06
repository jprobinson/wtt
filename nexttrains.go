package wtt

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/NYTimes/gizmo/server/kit"

	"github.com/jprobinson/gtfs/mta"
	"github.com/jprobinson/gtfs/transit_realtime"
)

func (s *service) getNextTrains(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*nextTrainsRequest)

	feed, err := GetFeed(ctx, s.hc, s.key, r.FeedType)
	if err != nil {
		kit.LogErrorMsg(ctx, err, "error getting feeds")
		return nil, kit.NewJSONStatusResponse(
			map[string]string{"error": "unable to get feed"},
			http.StatusInternalServerError)
	}

	alerts, north, south := mta.FeedNextTrainTimes(feed, r.Stop, r.Line)
	return &nextTrainResponse{
		Northbound: north,
		Southbound: south,
		Alerts:     alerts,
	}, nil
}

func decodeNextTrains(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := kit.Vars(r)
	line := strings.ToUpper(vars["line"])

	ft, err := parseFeed(line)
	if err != nil {
		return nil, err
	}
	return &nextTrainsRequest{
		FeedType: ft,
		Stop:     vars["stopID"],
		Line:     line,
	}, nil
}

type nextTrainsRequest struct {
	Stop     string
	Line     string
	FeedType mta.FeedType
}

type nextTrainResponse struct {
	Northbound []time.Time               `json:"northbound"`
	Southbound []time.Time               `json:"southbound"`
	Alerts     []*transit_realtime.Alert `json:"alerts"`
}
