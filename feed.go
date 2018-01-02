package main

import (
	"context"
	"strings"
	"time"

	"github.com/jprobinson/gosubway"
	"google.golang.org/appengine/log"
)

const (
	timeout     = 1 * time.Second
	maxAttempts = 10
	backoffStep = 50 * time.Millisecond
)

// retries until it hits max attempts or a context timeout
func getFeed(ctx context.Context, key string, ft gosubway.FeedType) (*gosubway.FeedMessage, error) {
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
		log.Errorf(ctx, "unable to get mta feed on attempt %d: %s", attempt, err)
	}
	return feed, err
}
