package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/nsqio/go-nsq"
)

type apiRequest struct {
	Addresses []string `json:"addresses"`
}

type apiResult struct {
	Results []VerfifyRes
}

func main() {

	// Initialize the nsq config
	cfg := nsq.NewConfig()
	cfg.Set("MaxInFlight", nsqMaxInFlight)

	// Create a Producer to send request to the workers
	producer, err := nsq.NewProducer(nsqdAddr, cfg)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/address/blocking", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got %q request to %q\n", r.Method, r.URL.Path)

		var apiReq apiRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&apiReq); err != nil {
			log.Printf("Decoding api request as json failed:%v\n", err)
			return
		}

		log.Printf("Got API req: +%v\n", apiReq)

		if len(apiReq.Addresses) == 0 {
			// Empty request
			return
		}

		// Create a NSQ consumer to pick up replies from the workers
		// Is this the most effective / best way to do this?
		// Does it connect once per http api call?

		// Create a random topic to use for the results - Based on nsq_tail
		rand.Seed(time.Now().UnixNano())
		resTopic := fmt.Sprintf(resTopicBase, rand.Int()%999999)
		consumer, err := nsq.NewConsumer(resTopic, managerChannel, cfg)
		if err != nil {
			log.Fatal(err)
		}

		// Handle the results
		consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
			log.Printf("Got result: %q\n", m.Body)
			return nil
		}))

		err = consumer.ConnectToNSQLookupd(nsqLookupAddr)
		if err != nil {
			log.Fatal(err)
		}

		for _, a := range apiReq.Addresses {

			// Make a request with the email and encode it as JSON
			req := VerfifyReq{Email: a, ResultTopic: resTopic}
			reqJSON, err := json.Marshal(req)
			if err != nil {
				log.Printf("api-handler-and-manger: Encoing result as JSON failed: %+v, err: %q\n", req, err)
				return //fmt.Errorf("api-handler-and-manger: Encoing result as JSON failed: %+v, err: %q\n", req, err)
			}

			// Publish to NSQ
			log.Printf("Publishing %q\t", req)
			producer.Publish(reqTopic, reqJSON)
			log.Println("done!")
		}

		// Loop indefinately and get the verify requests
		log.Println("LOOPING")
	FOO:
		for {
			select {
			case <-consumer.StopChan:
				//return
				break FOO
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(struct{ Foo string }{"BAR"}); err != nil {
			log.Println("Encoding JSON reply failed:", err)
			return
		}

	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
