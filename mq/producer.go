package mq

import (
	"filestore-server/config"
	"fmt"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

var notifyClose chan *amqp.Error

func init() {
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}

	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				fmt.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

func initChannel() bool {
	if channel != nil {
		return true
	}
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	channel, err = conn.Channel()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	return true
}

func Publish(exchange string, routingKey string, msg []byte) bool {
	if !initChannel() {
		return false
	}

	err := channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body: msg,
	})
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}