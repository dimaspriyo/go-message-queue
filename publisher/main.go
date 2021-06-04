package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"

	"github.com/bxcodec/faker/v3"
)

type Info struct {
	Name      string `json:"name" faker:"first_name"`
	IPv4      string `json:"ipv4" faker:"ipv4"`
	Byte      int64  `json:"byte" faker:"boundary_start=500, boundary_end=99999"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	e := echo.New()

	var conn *amqp.Connection
	var ch *amqp.Channel

	for {
		con, err := amqp.Dial("amqp://rabbitmq:5672")
		if err != nil {
			log.Println(err.Error())
		} else {
			channel, err := con.Channel()
			if err != nil {
				log.Println(err.Error())
			}

			err = channel.ExchangeDeclare(
				"appExchange", // name
				"fanout",      // type
				true,          // durable
				false,         // auto-deleted
				false,         // internal
				false,         // no-wait
				nil,           // arguments
			)
			if err != nil {
				log.Println(err.Error())
			}
			channel.Close()
			con.Close()
			break
		}
		time.Sleep(2 * time.Second)
	}

	conn, err := amqp.Dial("amqp://rabbitmq:5672")
	if err != nil {
		log.Println(err.Error())
	}
	defer conn.Close()

	ch, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"appExchange", // name
		"fanout",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Println(err.Error())
	}

	e.POST("", func(c echo.Context) error {

		info := Info{}
		err := faker.FakeData(&info)
		if err != nil {
			fmt.Println(err.Error())
		}

		info.Timestamp = time.Now().Unix()
		str, err := json.Marshal(&info)
		if err != nil {
			fmt.Println(err.Error())
		}

		body := string(str)
		err = ch.Publish(
			"appExchange", // exchange
			"",            // routing key
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(body),
			})
		if err != nil {
			panic(err.Error())
		}

		return nil
	})
	e.Logger.Fatal(e.Start(":1323"))
}
