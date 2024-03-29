package main

import (
	"encoding/json"
	"fmt"
	pubsub2 "github.com/cskr/pubsub"
	"github.com/streadway/amqp"
	"os"
)

var pubsub *amqp.Connection
var changesChannel *amqp.Channel
var changesQueue amqp.Queue

var localPubsub *pubsub2.PubSub

func connectToPubSub() {
	localPubsub = pubsub2.New(1)
	conn, err := amqp.Dial(os.Getenv("RABBIT"))
	if err != nil {
		fmt.Println(err)
		panic(err)
		return
	}
	pubsub = conn
	ch, err := pubsub.Channel()
	if err != nil {
		fmt.Println(err)
		panic(err)
		return
	}
	changesChannel = ch
	err = changesChannel.ExchangeDeclare("changes", "fanout", true, false, false, false, nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
		return
	}
	changesQueue, err = changesChannel.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
		return
	}
	err = changesChannel.QueueBind(changesQueue.Name, "", "changes", false, nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
		return
	}
	go func() {
		consumeChange(func(change resourceChange) {
			localPubsub.Pub(change, "changes")
		})
	}()
}

func publishChange(msg resourceChange) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = changesChannel.Publish("changes", "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(payload),
	})
	return err
}

type callback func(change resourceChange)

func consumeChange(callback callback) {
	msgs, err := changesChannel.Consume(changesQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		panic(err)
		return
	}
	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			var change resourceChange
			_ = json.Unmarshal(msg.Body, &change)
			callback(change)
		}
	}()
	<-forever
}
