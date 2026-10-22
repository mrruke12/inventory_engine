package rabbit

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(connString string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(connString)

		if err == nil {
			return conn, nil

		}
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

// func Consume(ch *amqp.Channel) error {
// 	// Prefetch: не тащить всё в память
// 	if err := ch.Qos(10, 0, false); err != nil {
// 		return err
// 	}

// 	deliveries, err := ch.Consume(
// 		"orders.q",
// 		"",    // consumer tag
// 		false, // autoAck = false — ОБЯЗАТЕЛЬНО для ack/nack
// 		false, // exclusive
// 		false, // noLocal
// 		false, // noWait
// 		nil,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	for d := range deliveries {
// 		if err := handle(d.Body); err != nil {
// 			// в DLQ через x-dead-letter-exchange
// 			_ = d.Nack(false, false) // multiple=false, requeue=false
// 			continue
// 		}
// 		_ = d.Ack(false)
// 	}
// 	return nil
// }
