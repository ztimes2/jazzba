package rabbit

import (
	"encoding/json"

	"github.com/ztimes2/jazzba/pkg/eventdriven"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// EventProducer implements evendriven.Producer interface and provides
// functionality for producing events to RabbitMQ.
type EventProducer struct {
	connection *amqp.Connection
}

// NewEventProducer initializes a new instance of EventProducer.
func NewEventProducer(conn *amqp.Connection) *EventProducer {
	return &EventProducer{
		connection: conn,
	}
}

// Produce produces an event message using a payload to RabbitMQ's queue
// associated with an event type.
func (ep EventProducer) Produce(eventType eventdriven.EventType,
	payload interface{}) error {

	queueName, ok := eventQueuesMap[eventType]
	if !ok {
		return newUnknownEventError(eventType)
	}

	channel, err := ep.connection.Channel()
	if err != nil {
		return errors.Wrap(err, "could not initialise amqp channel")
	}
	defer channel.Close()

	// Checks if a given queue exists.
	if _, err := channel.QueueDeclarePassive(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		return errors.Wrapf(err, "queue '%s' does not exist", queueName)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "could not marshal event payload")
	}

	if err := channel.Publish(
		// Empty exchange indicates the default exchange.
		"",
		// Using a queue name as routing key to publish a message directly to a
		// specific queue.
		queueName,
		// Mandatory flag is off since we don't handle cases when messages are
		// not delivered for now.
		false,
		// FIXME: Not sure what does immediate flag do exactly. Keeping if off
		// for now.
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  contentTypeJSON,
			Body:         body,
		},
	); err != nil {
		return errors.Wrap(err, "could not publish message to queue")
	}
	return nil
}
