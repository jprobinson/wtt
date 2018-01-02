package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/appengine/log"

	"github.com/NYTimes/marvin"
	"github.com/jprobinson/dialogflow"
	"github.com/jprobinson/gosubway"
)

func (s *service) postDialogflow(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*dialogflow.FulfillmentRequest)
	var (
		res string
		err error
	)
	log.Debugf(ctx, "request: %#v", r)
	switch r.Result.Action {
	case "next_train_request":
		line := r.Result.Parameters["subway-line"].(string)
		stop := r.Result.Parameters["subway-stop"].(string)
		dir := r.Result.Parameters["subway-direction"].(string)

		ft, err := parseFeed(line)
		if err != nil {
			log.Debugf(ctx, "unable to parse line: %s", line)
			return nil, err
		}

		res = s.getNextTrainDialog(ctx, ft, line, stop, dir)
	default:
		log.Debugf(ctx, "unkown action %s", r.Result.Action)
		return nil, errBadRequest
	}

	if err != nil {
		marvin.NewJSONStatusResponse(map[string]string{
			"error": "unable to complete request: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	log.Debugf(ctx, "responding with: %s", res)
	return &dialogflow.FulfillmentResponse{
		Speech:      res,
		DisplayText: res,
		Source:      "Where's The Train (NYC)",
	}, nil
}

var errBadRequest = marvin.NewJSONStatusResponse(map[string]string{
	"error": "bad request"}, http.StatusBadRequest)

func decodeDialogflow(ctx context.Context, r *http.Request) (interface{}, error) {
	var req dialogflow.FulfillmentRequest
	bod, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Debugf(ctx, "unable to read request: %s", err)
		return nil, errBadRequest
	}
	err = json.Unmarshal(bod, &req)
	if err != nil {
		log.Debugf(ctx, "unable to decode request: %s - %s", err, string(bod))
		return nil, errBadRequest
	}
	defer r.Body.Close()
	return &req, nil
}

func (s *service) getNextTrainDialog(ctx context.Context, ft gosubway.FeedType, line, stop, dir string) string {

	feed, err := getFeed(ctx, s.key, ft)
	if err != nil {
		return fmt.Sprintf("Sorry, I'm having problems getting the subway feed. Please try again later. \"%s\"", stop)
	}

	stopLine, ok := stopNameToID[stop]
	if !ok {
		return fmt.Sprintf("Sorry, I didn't recognise the stop \"%s\"", stop)
	}

	stopID, ok := stopLine[line]
	if !ok {
		return fmt.Sprintf("Sorry, I didn't recognise \"%s\" as a part of the %s line",
			stop, line)
	}

	_, north, south := feed.NextTrainTimes(stopID, line)

	var trains []time.Time
	if trainDirs[line]["northbound"] == dir {
		trains = north
	} else {
		trains = south
	}

	if len(trains) == 0 {
		return fmt.Sprintf("Sorry, there are no train times available for %s bound %s trains at %s",
			dir, line, stop)
	}

	diff := trains[0].Sub(time.Now().UTC())
	mins := strconv.Itoa(int(diff.Minutes()))
	secs := strconv.Itoa(int(diff.Seconds()) % 60)
	out := "The next train will leave in "
	if mins != "0" {
		out += mins + " minutes and "
	}
	out += secs + " seconds"
	return out
}
