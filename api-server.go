package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"os"
	"path/filepath"
)

// Incoming REST API requests
type apiRequest struct {
	Addresses []string `json:"addresses"`
	Callback  string   `json:"callback"` // Will be empty for /address/blocking
}

// Outgoing REST API Reply, for calls to /address/blocking
type apiReplyBlocking struct {
	Results []VerifyRes // FIXME: Include the other type
}

// Outgoing REST API Reply, for calls to /address/callback
type apiReplyCallback struct {
	StatusOK bool   `json:"status-ok"`
	Message  string `json:"message"`
}

// The actual results of running a job,
// delivered either in the response to calls to /address/blocking or as callbacks when using /address/callback
type apiResult struct {
	Results []VerifyRes
}

var producer *nsq.Producer // FIXME ??
var nsqLookupdHost string

func main() {

	// Check arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage:\n\t", filepath.Base(os.Args[0]), " nsqd host  nsqlookupd host\n(Only 1 lookupd supported atm, fixme)")
		os.Exit(0)
	}
	nsqdHost := os.Args[1]
	nsqLookupdHost = os.Args[2]

	// Initialize the nsq config
	cfg := nsq.NewConfig()

	// Create a Producer to send request to the workers
	var err error
	producer, err = nsq.NewProducer(nsqdHost+":"+nsqdPort, cfg)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/address/blocking", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got %q request to %q\n", r.Method, r.URL.Path)

		// Decode the JSON request - FIXME: duplicated code
		var apiReq apiRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&apiReq); err != nil {
			log.Printf("Decoding api request as json failed:%v\n", err)
			return
		}

		log.Printf("Got API req: +%v\n", apiReq)

		var result apiReplyBlocking
		result.Results, err = runJob(apiReq) // this blocks.. for a long time
		if err != nil {
			log.Printf("Failed running Job: %q\n", err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Println("Encoding JSON reply failed:", err)
			return
		}

	})

	http.HandleFunc("/address/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got %q request to %q\n", r.Method, r.URL.Path)

		// Decode the JSON request
		var apiReq apiRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&apiReq); err != nil {
			log.Printf("Decoding api request as json failed:%v\n", err)
			return
		}

		log.Printf("Got API req: +%v\n", apiReq)
		if apiReq.Callback == "" {
			log.Printf("Inavlid API Request, missing callback: %+v", apiReq)
			return
		}

		go func() {
			// Send the requests to the workers and wait for the results
			var result apiResult
			result.Results, err = runJob(apiReq) // this blocks.. for a long time
			if err != nil {
				log.Printf("Failed running Job: %q\n", err)
				return
			}

			// Post the results back to the callback url
			log.Printf("POSTING:\n%+v\n\nTo:\n%q\n", result, apiReq.Callback)
		}()

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(apiReplyCallback{StatusOK: true, Message: "Results will be POSTed to: " + apiReq.Callback}); err != nil {
			log.Println("Encoding JSON reply failed:", err)
			return
		}

	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func runJob(apiReq apiRequest) ([]VerifyRes, error) {

	if len(apiReq.Addresses) == 0 {
		// Empty request
		return nil, nil
	}

	// Create a NSQ consumer to pick up replies from the workers
	// Is this the most effective / best way to do this?
	// Does it connect once per http api call?

	// Initialize the nsq config
	cfg := nsq.NewConfig()

	// Create a random topic to use for the results - Based on nsq_tail
	rand.Seed(time.Now().UnixNano())
	resTopic := fmt.Sprintf(resTopicBase, rand.Int()%999999)
	consumer, err := nsq.NewConsumer(resTopic, managerChannel, cfg)
	if err != nil {
		log.Fatal(err)
	}

	consumer.ChangeMaxInFlight(nsqMaxInFlight) // cfg.Set() did not work for some reason

	// Make sure we gather all responses.. or timeout
	// Nice example: http://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
	wgResults := sync.WaitGroup{}
	var result []VerifyRes

	// Handle the results
	consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		var res VerifyRes
		err := json.Unmarshal(m.Body, &res)
		if err != nil {
			// log.Printf("Decoding verify result as json failed:%v\n", err)
			return fmt.Errorf("Decoding verify result as json failed: %v\n", err)
		}

		log.Printf("Got result: %q\n", m.Body)
		result = append(result, res)
		wgResults.Done()
		return nil
	}))

	err = consumer.ConnectToNSQLookupd(nsqLookupdHost + ":" + nsqLookupdPort)
	if err != nil {
		log.Fatal(err)
	}

	for _, a := range apiReq.Addresses {

		// Make a request with the email and encode it as JSON
		req := VerifyReq{Email: a, ResultTopic: resTopic}
		reqJSON, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("api-handler-and-manger: Encoing result as JSON failed: %+v, err: %q\n", req, err)
			// return //fmt.Errorf("api-handler-and-manger: Encoing result as JSON failed: %+v, err: %q\n", req, err)
		}

		// Publish to NSQ
		// MultiPublish instead?
		log.Printf("Publishing %q\t", req)
		producer.Publish(reqTopic, reqJSON)
		log.Println("done!")

		wgResults.Add(1)
	}

	done := make(chan bool)
	go func() {
		defer close(done) // Use close as a signal
		wgResults.Wait()
	}()

	select {
	case <-consumer.StopChan:
		//return
		log.Println("Consumer Stopping")
	case <-done:
		log.Println("Job Done!")
		consumer.Stop()
	case <-time.After(time.Minute * 3):
		log.Println("Jobe Timeout!")
		consumer.Stop()
	}

	return result, nil

}
