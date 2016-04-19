package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nsqio/go-nsq"
	"github.com/oyvindsk/go-email-address-check/emailVerify"
)

// A global producer, so it's easy to access from all function.
// FIXME Should not be this global, make a type for the handle and put it there?
var producer *nsq.Producer

func main() {

	// Initialize the nsq config
	cfg := nsq.NewConfig()
	cfg.Set("MaxInFlight", nsqMaxInFlight)

	// Create a consumer to pick up new verify requests
	consumer, err := nsq.NewConsumer(reqTopic, workerChannel, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// handle a Lookup Request message
	consumer.AddHandler(nsq.HandlerFunc(handleVerifyRequest))

	// Create a Producer to send the responses
	producer, err = nsq.NewProducer(nsqdAddr, cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = consumer.ConnectToNSQLookupd(nsqLookupAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Loop indefinately and get the verify requests
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
	var res VerfifyRes

	// decode the json message
	var req VerfifyReq
	err := json.Unmarshal(m.Body, &req)
	if err != nil {
		log.Printf("verify-address-nsq: Saw invalid json: %q, err: %q\n", m.Body, err)
		res.Error = fmt.Errorf("verify-address-nsq: Saw invalid json: %q, err: %q\n", m.Body, err)
	}

	if res.Error == nil {

		res.Email = req.Email

		// check the address
		addrOK, smtpMsg, err := emailVerify.VerifyAddress(req.Email) // FIXME ? Use []byte for lib as well
		if err != nil {
			log.Printf("verify-address-nsq: Looking up adr: %+v failed: %q\n", req, err)
			res.Error = fmt.Errorf("verify-address-nsq: Looking up adr: %+v failed: %q\n", req, err)
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
