package wtt

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/NYTimes/gizmo/server/kit"
	"github.com/jprobinson/gosubway"
)

const (
	timeout     = 1 * time.Second
	maxAttempts = 10
	backoffStep = 50 * time.Millisecond
)

// retries until it hits max attempts or a context timeout
func GetFeed(ctx context.Context, key string, ft gosubway.FeedType) (*gosubway.FeedMessage, error) {
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
		feed, err = gosubway.GetFeed(ctx, key, ft)
		if err == nil ||
			(err != nil && strings.Contains(err.Error(), "deadline exceeded")) {
			break
		}
		kit.LogErrorMsg(ctx, err, "unable to get mta feed")
	}
	return feed, err
}

func parseFeed(line string) (gosubway.FeedType, error) {
	var ft gosubway.FeedType
	switch line {
	case "1", "2", "3", "4", "5", "6":
		ft = gosubway.NumberedFeed
	case "N", "Q", "R", "W":
		ft = gosubway.YellowFeed
	case "B", "D", "F", "M":
		ft = gosubway.OrangeFeed
	case "A", "C", "E":
		ft = gosubway.BlueFeed
	case "J", "Z":
		ft = gosubway.BrownFeed
	case "L":
		ft = gosubway.LFeed
	case "7":
		ft = gosubway.SevenFeed
	case "G":
		ft = gosubway.GFeed
	default:
		return gosubway.LFeed, errBadRequest
	}
	return ft, nil
}

var errBadRequest = kit.NewJSONStatusResponse(map[string]string{
	"error": "bad request"}, http.StatusBadRequest)
