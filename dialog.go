package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine/datastore"
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
	switch r.Result.Action {
	case "my_next_train_request":
		uid := r.OriginalRequest.Data.User.UserID
		if uid == "" {
			res = "sorry, you need to be logged in for that to work"
			break
		}
		mys, serr := getMyStop(ctx, uid)
		if serr == datastore.ErrNoSuchEntity {
			res = "It looks like you haven't saved your personalized subway stop yet! Ask NYC Train Time to \"save my stop\" to create or update your stop."
			break
		}
		if serr != nil {
			err = serr
			res = "sorry, we were unable to look up your stop."
			break
		}
		ft, err := parseFeed(mys.Line)
		if err != nil {
			log.Debugf(ctx, "unable to parse line: %s", mys.Line)
			res = fmt.Sprintf("sorry, the %s line is not available yet", mys.Line)
			break
		}
		res = s.getNextTrainDialog(ctx, ft, mys.Line, mys.Stop, mys.Dir)
	case "my_following_train_request":
		uid := r.OriginalRequest.Data.User.UserID
		if uid == "" {
			res = "Sorry, you need to be logged in for that to work."
			break
		}
		mys, serr := getMyStop(ctx, uid)
		if serr == datastore.ErrNoSuchEntity {
			res = "You haven't saved your personalized subway stop yet. Ask NYC Train Time to \"save my stop\" to create or update your stop. "
			break
		}
		if serr != nil {
			err = serr
			res = "sorry, we were unable to look up your stop."
			break
		}
		ft, err := parseFeed(mys.Line)
		if err != nil {
			log.Debugf(ctx, "unable to parse line: %s", mys.Line)
			res = fmt.Sprintf("sorry, the %s line is not available yet. ", mys.Line)
			break
		}
		res = s.getFollowingTrainDialog(ctx, ft, mys.Line, mys.Stop, mys.Dir)

	case "save_my_stop_request":
		uid := r.OriginalRequest.Data.User.UserID
		if uid == "" {
			res = "Sorry, you need to be logged in for that to work"
			break
		}
		line := strings.ToUpper(r.Result.Parameters["subway-line"].(string))
		stop := r.Result.Parameters["subway-stop"].(string)
		dir := r.Result.Parameters["subway-direction"].(string)

		err = saveMyStop(ctx, uid, line, stop, dir)
		if err != nil {
			return nil, marvin.NewJSONStatusResponse(map[string]string{
				"error": "unable to complete request: " + err.Error(),
			}, http.StatusInternalServerError)
		}
		res = fmt.Sprintf(
			"Successfully saved your stop, %s bound %s trains at %s. To update your stop again, ask NYC Train Time to \"save my stop\". ",
			dir, line, stop)
	case "next_train_request":
		line := strings.ToUpper(r.Result.Parameters["subway-line"].(string))
		stop := r.Result.Parameters["subway-stop"].(string)
		dir := r.Result.Parameters["subway-direction"].(string)

		ft, err := parseFeed(line)
		if err != nil {
			log.Debugf(ctx, "unable to parse line: %s", line)
			res = fmt.Sprintf("sorry, the %s line is not available yet", line)
			break
		}

		res = s.getNextTrainDialog(ctx, ft, line, stop, dir) +
			" If you would like me to remember your stop, ask NYC Train Time to \"save my stop\" and then ask for MY stop next time. "
	case "following_train_request":
		line := strings.ToUpper(r.Result.Parameters["subway-line"].(string))
		stop := r.Result.Parameters["subway-stop"].(string)
		dir := r.Result.Parameters["subway-direction"].(string)
		ft, err := parseFeed(line)
		if err != nil {
			log.Debugf(ctx, "unable to parse line: %s", line)
			res = fmt.Sprintf("Sorry, the %s line is not available yet.", line)
			break
		}
		res = s.getFollowingTrainDialog(ctx, ft, line, stop, dir)
	default:
		log.Debugf(ctx, "unkown action %s", r.Result.Action)
		return nil, errBadRequest
	}

	if err != nil {
		return nil, marvin.NewJSONStatusResponse(map[string]string{
			"error": "unable to complete request: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	// random goodbye
	res += " ..." + goodbyes[rand.New(rand.NewSource(time.Now().Unix())).Intn(len(goodbyes)-1)]

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
	log.Debugf(ctx, "bod: %s", string(bod))
	err = json.Unmarshal(bod, &req)
	if err != nil {
		log.Debugf(ctx, "unable to decode request: %s - %s", err, string(bod))
		return nil, errBadRequest
	}
	defer r.Body.Close()
	return &req, nil
}

func (s *service) getNextTrainDialog(ctx context.Context, ft gosubway.FeedType, line, stop, dir string) string {
	return s.getTrainDialog(ctx, ft, "next", line, stop, dir, 0)
}

func (s *service) getFollowingTrainDialog(ctx context.Context, ft gosubway.FeedType, line, stop, dir string) string {
	return s.getTrainDialog(ctx, ft, "following", line, stop, dir, 1)
}

func (s *service) getTrainDialog(ctx context.Context, ft gosubway.FeedType, name, line, stop, dir string, indx int) string {
	feed, err := getFeed(ctx, s.key, ft)
	if err != nil {
		return fmt.Sprintf("Sorry, I'm having problems getting the subway feed. ")
	}

	stopLine, ok := stopNameToID[stop]
	if !ok {
		return fmt.Sprintf("Sorry, I didn't recognise the stop \"%s\". ", stop)
	}

	stopID, ok := stopLine[line]
	if !ok {
		return fmt.Sprintf("Sorry, I didn't recognise \"%s\" as a part of the %s line. ",
			stop, line)
	}

	_, north, south := feed.NextTrainTimes(stopID, line)

	var trains []time.Time
	if trainDirs[line]["northbound"] == dir || dir == "uptown" || dir == "Northbound" {
		trains = north
	} else {
		trains = south
	}

	if len(trains) < indx+1 {
		return fmt.Sprintf("Sorry, the %s train time is not available for %s bound %s trains at %s. ",
			name, dir, line, stop)
	}

	out := timeSpeak(trains[indx], name, line, stop, dir)
	if len(trains) >= indx+2 {
		out += timeSpeak(trains[indx+1], "following", line, stop, dir)
	}
	return out
}

func timeSpeak(t time.Time, name, line, stop, dir string) string {
	diff := t.Sub(time.Now().UTC())
	mins := strconv.Itoa(int(diff.Minutes()))
	secs := strconv.Itoa(int(diff.Seconds()) % 60)
	out := fmt.Sprintf("The %s %s train will leave %s towards %s in ",
		name, line, stop, dir)
	if mins != "0" {
		out += mins + " minutes and "
	}
	out += secs + " seconds. "
	return out
}

type myStop struct {
	Line string
	Stop string
	Dir  string
}

func getMyStop(ctx context.Context, userID string) (*myStop, error) {
	var my myStop
	err := datastore.Get(ctx, datastore.NewKey(ctx, "MyStop", userID, 0, nil), &my)
	return &my, err
}

func saveMyStop(ctx context.Context, userID, line, stop, dir string) error {
	_, err := datastore.Put(ctx, datastore.NewKey(ctx, "MyStop", userID, 0, nil), &myStop{
		Line: line,
		Stop: stop,
		Dir:  dir,
	})
	return err
}

var goodbyes = []string{
	"Ok, bye!",
	"Bye bye now",
	"Peace out!",
	"Goodbye",
	"Hope you can catch the train!",
	"Hope you can make it!",
	"Adios!",
	"Au revoir",
	"Have a good trip!",
	"Have a good ride!",
	"Have a save trip!",
	"Save travels!",
}
