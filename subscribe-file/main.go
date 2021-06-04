package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	for {
		conn, err := amqp.Dial("amqp://rabbitmq:5672")
		if err != nil {
			log.Println(err.Error())
		} else {
			for {
				ch, err := conn.Channel()
				if err != nil {
					log.Println(err.Error())
				} else {
					ch.Close()
					break
				}
			}
			conn.Close()
			break
		}
		time.Sleep(2 * time.Second)
	}

	conn, err := amqp.Dial("amqp://rabbitmq:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"subscribe-file-queue", // name
		false,                  // durable
		false,                  // delete when unused
		true,                   // exclusive
		false,                  // no-wait
		nil,                    // arguments
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

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
			file, err := os.OpenFile("activity.log", os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(err.Error())
			}
			defer file.Close()

			file.Write(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
