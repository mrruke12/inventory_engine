package rabbit

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type DeclareTopologyFunc func(conn *amqp.Connection) error

func DeclareTopology(conn *amqp.Connection) error {
	return nil // TODO: define topology
}
