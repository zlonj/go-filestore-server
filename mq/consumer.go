package mq

import "fmt"

var done chan bool

func StartConsume(queueName string, consumerName string, callback func(msg []byte) bool) {
	msgs, err := channel.Consume(queueName, consumerName, true, false, false, false, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	done = make(chan bool)

	go func() {
		for msg := range msgs {
			processSuccess := callback(msg.Body)
			if !processSuccess {
				// TODO: Insert msg to error queue for retries
				fmt.Println("Retry")
			}
		}
	}()

	<- done
	channel.Close()
}