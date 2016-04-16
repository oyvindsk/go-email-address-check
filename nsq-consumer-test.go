package main

import (
	"fmt"
	"log"

	"github.com/nsqio/go-nsq"
)

func main() {

	cfg := nsq.NewConfig()

	consumer, err := nsq.NewConsumer("foo2", "ch1", cfg)
	if err != nil {
		log.Fatal(err)
	}

	consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		// handle the message
		fmt.Printf("Got message:%q", m.Body)
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
