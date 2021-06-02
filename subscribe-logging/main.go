package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
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

	file, err := os.OpenFile("application.log", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	conn, err := amqp.Dial("amqp://localhost:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"subscribe-logging-queue", // name
		false,                     // durable
		false,                     // delete when unused
		true,                      // exclusive
		false,                     // no-wait
		nil,                       // arguments
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
			log.Printf(" [x] %s", d.Body)
			err = json.Unmarshal(d.Body, &data)
			if err != nil {
				panic(err.Error())
			}

			_, err = file.WriteString(fmt.Sprintf("\n%s##%s##%d##%d", data.Name, data.IPv4, data.Byte, data.Timestamp))
			if err != nil {
				panic(err.Error())
			}
		}
	}()
	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
