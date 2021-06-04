package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
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

	fluentdConfig := fluent.Config{
		FluentHost: "fluentd",
		FluentPort: 24224,
	}

	for {
		logger, err := fluent.New(fluentdConfig)
		if err != nil {
			log.Println(err.Error())
		} else {
			logger.Close()
			break
		}
		time.Sleep(2 * time.Second)
	}

	logger, err := fluent.New(fluentdConfig)
	if err != nil {
		log.Println(err.Error())
	}
	defer logger.Close()

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
			err = json.Unmarshal(d.Body, &data)
			if err != nil {
				log.Println(err.Error())
			}
			log.Printf(" [x] %d", data.Byte)

			var data = map[string]int64{
				"byte": data.Byte,
			}

			e := logger.Post("go-subscribe-logging.log", data)
			if e != nil {
				log.Println("Error while posting log: ", e)
			} else {
				log.Println("Success to post log")
			}
		}
	}()
	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
