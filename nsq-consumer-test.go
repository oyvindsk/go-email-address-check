package main

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"log"
)

func main() {

	cfg := nsq.NewConfig()

	consumer, err := nsq.NewConsumer("foo", "ch1", cfg)
	if err != nil {
		log.Fatal(err)
	}

	consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		// handle the message
		fmt.Println("Got message:", string(m.Body))
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
