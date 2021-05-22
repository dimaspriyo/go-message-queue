package main

import (
	"encoding/json"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"

	"github.com/bxcodec/faker/v3"
)

type Info struct {
	Name      string `json:"name" faker:"first_name"`
	IPv4      string `json:"ipv4" faker:"ipv4"`
	Byte      int64  `json:"byte" faker:"boundary_start=500, boundary_end=99999"`
	Timestamp string `json:"timestamp" faker:"timestamp"`
}

func main() {
	e := echo.New()
	e.POST("", func(c echo.Context) error {
		conn, err := amqp.Dial("amqp://localhost:5672")
		if err != nil {
			panic(err.Error())
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			panic(err.Error())
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

		info := Info{}
		err = faker.FakeData(&info)
		if err != nil {
			fmt.Println(err.Error())
		}

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
