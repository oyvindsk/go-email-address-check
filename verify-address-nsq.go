package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nsqio/go-nsq"
	"github.com/oyvindsk/go-email-address-check/emailVerify"
)

const (
	reqTopic   = "verify-requests"
	resTopic   = "verify-results"
	reqChannel = "workers"

	nsqdAddr      = "127.0.0.1:4150"
	nsqLookupAddr = "127.0.0.1:4160"
)

// VerfifyReq is the type used for when a verfification request goes through NSQ
type VerfifyReq struct {
	Email string
}

// VerfifyRes is the type used for when a verfification response goes through NSQ
type VerfifyRes struct {
	Email     string
	AddressOK bool
	SMTPMsg   string
	Error     error
}

func main() {

	// Initialize the nsq config
	cfg := nsq.NewConfig()

	// Create a consumer to pick up new verify requests
	consumer, err := nsq.NewConsumer(reqTopic, reqChannel, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create a Producer to send the responses
	producer, err := nsq.NewProducer(nsqdAddr, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// handle a Lookup Request message
	consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		log.Printf("Got message: %q\n", m.Body)

		// do some basic sanity checking
		//if len(m.Body) == 0 {
		//	log.Printf("verify-address-nsq: Saw invalid req: %q\n", m.Body)
		//	return fmt.Errorf("verify-address-nsq: Saw invalid req: %q", m.Body)
		//}

		// decode the json message
		var req VerfifyReq
		err = json.Unmarshal(m.Body, &req)
		if err != nil {
			log.Printf("verify-address-nsq: Saw invalid json: %q, err: %q\n", m.Body, err)
			return fmt.Errorf("verify-address-nsq: Saw invalid json: %q, err: %q\n", m.Body, err)
		}

		// check the address
		addrOK, smtpMsg, err := emailVerify.VerifyAddress(req.Email) // FIXME ? Use []byte for lib as well

		if err != nil {
			log.Printf("verify-address-nsq: Looking up adr: %+v failed: %q\n", req, err)
			return fmt.Errorf("verify-address-nsq: Looking up adr: %+v failed: %q\n", req, err)

		}

		// Prepare the response
		res := VerfifyRes{Email: req.Email, AddressOK: addrOK, SMTPMsg: smtpMsg, Error: err}
		resJSON, err := json.Marshal(res)
		if err != nil {
			log.Printf("verify-address-nsq: Encoing result as JSON failed: %+v, err: %q\n", req, err)
			return fmt.Errorf("verify-address-nsq: Encoing result as JSON failed: %+v, err: %q\n", req, err)
		}

		// Send the response
		err = producer.Publish(resTopic, resJSON)
		if err != nil {
			log.Printf("verify-address-nsq: Publishing result to NSQ failed: %+v, err: %q\n", resJSON, err)
			return fmt.Errorf("verify-address-nsq: Publishing result to NSQ failed: %+v, err: %q\n", resJSON, err)
		}

		if addrOK {
			log.Printf("OK, Address %+v seems valid (smtp msg: %q)", req, smtpMsg)
		} else {
			log.Printf("Invalid, Address %+v seems invalid (smtp msg: %q)", req, smtpMsg)
		}

		return nil

	}))

	err = consumer.ConnectToNSQLookupds([]string{"127.0.0.1:4161"})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-consumer.StopChan:
			return
		}
	}

}
