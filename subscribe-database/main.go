package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Info struct {
	Name      string `json:"name"`
	IPv4      string `json:"ipv4"`
	Byte      int64  `json:"byte"`
	Timestamp int64  `json:"timestamp"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:root@localhost:27017"))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	collection := client.Database("go-message-queue").Collection("subscribe-database")

	conn, err := amqp.Dial("amqp://localhost:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"appQueue", // name
		false,      // durable
		false,      // delete when unused
		true,       // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,        // queue name
		"",            // routing key
		"appExchange", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	var data Info
	go func() {
		for d := range msgs {

			err := json.Unmarshal(d.Body, &data)
			if err != nil {
				panic(err.Error())
			}
			_, err = collection.InsertOne(ctx, bson.D{
				{Key: "name", Value: data.Name},
				{Key: "ipv4", Value: data.IPv4},
				{Key: "byte", Value: data.Byte},
				{Key: "timestamp", Value: data.Timestamp}})

			if err != nil {
				panic(err.Error())
			}

			fmt.Printf("Insert Name=%s, IPV4=%s, Byte=%d, Timestamp=%d", data.Name, data.IPv4, data.Byte, data.Timestamp)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
