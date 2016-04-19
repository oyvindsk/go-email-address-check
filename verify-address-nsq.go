package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/oyvindsk/go-email-address-check/emailVerify"
)

const (
	reqTopic = "verify-requests"
	resTopic = "verify-results"
)

func main() {

	// Make up a channel name
	channel := ""
	if channel == "" {
		rand.Seed(time.Now().UnixNano())
		channel = fmt.Sprintf("tail%06d#ephemeral", rand.Int()%999999)
	}

	// Initialize the nsq config
	cfg := nsq.NewConfig()

	// Create a consumer to pick up new verify requests
	consumer, err := nsq.NewConsumer(reqTopic, channel, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// handle a Lookup Request message
	consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		log.Printf("Got message: %q\n", m.Body)

		if len(m.Body) == 0 {
			log.Printf("verify-address-nsq: Saw invalid req: %q\n", m.Body)
		}

		addrOK, smtpMsg, err := emailVerify.VerifyAddress(string(m.Body)) // FIXME ? Use []byte for lib as well

		if err != nil {
			log.Printf("verify-address-nsq: Looking up adr: %q failed: %q\n", m.Body, err)
		}

		if addrOK {
			log.Printf("OK, Address %q seems valid (smtp msg: %q)", m.Body, smtpMsg)
		} else {
			log.Printf("Invalid, Address %q seems invalid (smtp msg: %q)", m.Body, smtpMsg)
		}

		return nil // FIXME
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
