package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// PathAPIURL is from https://www.reddit.com/r/jerseycity/comments/bb4041/programmatic_realtime_path_data/
const PathAPIURL = "https://path.api.razza.dev/v1/stations/fourteenth_street/realtime"

// CodeCard data type
//https://github.com/cameronsenese/codecard/tree/master/functions#create-a-fn-function-for-your-code-card
// {
//	"template": "template[1-11]",
//	"title": "Hello World",
//	"subtitle": "This is a subtitle",
//	"bodytext": "This is the body", (The document above has typo)
//	"icon": "[see list of named icons| BMP url]",
//	"backgroundColor": "[white|black]"
//}
// https://github.com/cameronsenese/codecard/blob/master/arduino/codecard/dataParser.h
type CodeCard struct {
	Template        string `json:"template"`
	Title           string `json:"title"`
	Subtitle        string `json:"subtitle"`
	BodyText        string `json:"bodytext"`
	Icon            string `json:"icon"`
	BackgroundColor string `json:"backgroundColor"`
}

// PathUpcomingTrain is from https://path.api.razza.dev/v1/stations/fourteenth_street/realtime
// {"lineName":"33rd Street via Hoboken","lineColors":["#4D92FB","#FF9900"],
// "projectedArrival":"2019-09-22T02:51:11Z","lastUpdated":"2019-09-22T02:46:51Z",
// "status":"ON_TIME","headsign":"33rd Street via Hoboken",
// "route":"JSQ_33_HOB","routeDisplayName":"Journal Square - 33rd Street (via Hoboken)",
// "direction":"TO_NY"}
type PathUpcomingTrain struct {
	LineName         string    `json:"lineName"`
	LineColors       []string  `json:"lineColors"`
	ProjectedArrival time.Time `json:"projectedArrival"`
	LastUpdated      time.Time `json:"lastUpdated"`
	Status           string    `json:"status"`
	Route            string    `json:"route"`
	RouteDisplayName string    `json:"routeDisplayName"`
	Direction        string    `json:"direction"`
}

// PathResponse is from https://path.api.razza.dev/v1/stations/fourteenth_street/realtime
type PathResponse struct {
	UpcomingTrains []PathUpcomingTrain `json:"upcomingTrains"`
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("Received a request.")

	// CodeCard rejects response with HTTP/1.0 even though the request is HTTP/1.0
	// https://github.com/cameronsenese/codecard/commit/58880db2f32c391abce28eda90500a4e98580d80
	// It turned out that Cloud Run's HTTPS front-end automatically converts this
	// HTTP version. Thus this does not solve the problem. I had to upgrade CodeCard firmware.
	r.Proto = "HTTP/1.1"
	r.ProtoMajor = 1
	r.ProtoMinor = 1

	var request []string // Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)                             // Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host)) // Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}
	log.Print(strings.Join(request, "\n"))

	pathResponse, err := myClient.Get(PathAPIURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer pathResponse.Body.Close()

	pathResponseJSON := new(PathResponse)
	json.NewDecoder(pathResponse.Body).Decode(&pathResponseJSON)

	log.Print(pathResponseJSON)

	currentTime := time.Now()
	var upcomingTrainDuration []time.Duration
	bodyText := ""
	var status string
	for _, upcomingTrain := range pathResponseJSON.UpcomingTrains {
		if !strings.Contains(upcomingTrain.Route, "JSQ") || upcomingTrain.Direction != "TO_NJ" {
			continue
		}
		status = upcomingTrain.Status
		timeDiff := upcomingTrain.ProjectedArrival.Sub(currentTime)
		upcomingTrainDuration = append(upcomingTrainDuration, timeDiff)
		diffMinutes := timeDiff.Round(time.Minute) / time.Minute
		bodyText = fmt.Sprintf("%s%d minutes\n", bodyText, diffMinutes)
	}

	codeCard := CodeCard{"template1", "Hello CodeCard", "JSQ train: " + status, bodyText, "01d",
		"white"}
	codeCardJSON, err := json.Marshal(codeCard)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(codeCardJSON)
}

func main() {
	log.Print("Hello world sample started.")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
