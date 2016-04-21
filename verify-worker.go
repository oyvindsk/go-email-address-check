package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nsqio/go-nsq"
	"github.com/oyvindsk/go-email-address-check/verify"
	"os"
	"path/filepath"
	"time"
)

const (
	maxConcurrentChecks = 50
)

// A global producer, so it's easy to access from all function.
// FIXME Should not be this global, make a type for the handle and put it there?
// Is it OK that many go routines publishes to the same producer?
var producer *nsq.Producer


type Footype struct {
	foo bool
}

//	func (h *Foo) HandleMessage(m bool) error {

func (th *Footype) HandleMessage(m *nsq.Message) error {
	log.Println("handler 2!")
	time.Sleep(5 * time.Second)
	log.Println("handler Done!")
	return nil
}

func main() {

	// Check arguments
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n\t", filepath.Base(os.Args[0]), " nsqd host\n(Just 1 nsqd, so no HA atm, fixme)")
		os.Exit(0)
	}
	nsqdHost := os.Args[1]

	// Initialize the nsq config
	cfg := nsq.NewConfig()

	// Create a consumer to pick up new verify requests
	consumer, err := nsq.NewConsumer(reqTopic, workerChannel, cfg)
	if err != nil {
		log.Fatal(err)
	}

	consumer.ChangeMaxInFlight(nsqMaxInFlight) // cfg.Set() did not work for some reason

	//consumer.SetLogger(log.New(os.Stdout, "!! ", log.Lshortfile), nsq.LogLevelDebug)

	// handle a Verify Request message
	consumer.AddConcurrentHandlers(nsq.HandlerFunc(handleVerifyRequest), 10)

	// Create a Producer to send the responses
	producer, err = nsq.NewProducer(nsqdHost+":"+nsqdPort, cfg)
	if err != nil {
		log.Fatal(err)
	}

	//err = consumer.ConnectToNSQLookupd(nsqLookupdHost + ":" + nsqLookupdPort)
	err = consumer.ConnectToNSQD(nsqdHost + ":" + nsqdPort)
	if err != nil {
		log.Fatal(err)
	}

	// Loop indefinitely and get the verify requests
	for {
		select {
		case <-consumer.StopChan:
			return
		}
	}

}

// handleVerifyRequest runs once for each verify request seen on the NSQ topic
// one go routine per request (message) ??
func handleVerifyRequest(m *nsq.Message) error {
	log.Printf("Got message: %q\n", m.Body)

	// The result to publish to NSQ
	var res VerifyRes

	// decode the json message
	var req VerifyReq
	err := json.Unmarshal(m.Body, &req)
	if err != nil {
		log.Printf("verify-address-nsq: Saw invalid json: %q, err: %q\n", m.Body, err)
		res.Error = fmt.Sprintf("verify-address-nsq: Saw invalid json: %q, err: %q\n", m.Body, err)
	}

	if res.Error == "" {

		res.Email = req.Email

		// check the address
		addrOK, smtpMsg, err := verify.VerifyAddress(req.Email) // FIXME ? Use []byte for lib as well
		if err != nil {
			log.Printf("verify-address-nsq: Looking up adr: %+v failed: %q\n", req, err)
			res.Error = fmt.Sprintf("verify-address-nsq: Looking up adr: %+v failed: %q\n", req, err)
		} else {
			res.AddressOK = addrOK
			res.SMTPMsg = smtpMsg
		}
	}

	// Prepare the response
	resJSON, err := json.Marshal(res)
	if err != nil {
		log.Printf("verify-address-nsq: Encoing result as JSON failed: %+v, err: %q\n", req, err)
		return fmt.Errorf("verify-address-nsq: Encoing result as JSON failed: %+v, err: %q\n", req, err)
	}

	// Send the response
	err = producer.Publish(req.ResultTopic, resJSON)
	if err != nil {
		log.Printf("verify-address-nsq: Publishing result to NSQ failed: %+v, err: %q\n", resJSON, err)
		return fmt.Errorf("verify-address-nsq: Publishing result to NSQ failed: %+v, err: %q\n", resJSON, err)
	}

	if res.AddressOK {
		log.Printf("OK, Address %+v seems valid (smtp msg: %q)", req, res.SMTPMsg)
	} else {
		log.Printf("Invalid, Address %+v seems invalid (smtp msg: %q)", req, res.SMTPMsg)
	}

	return nil

}
