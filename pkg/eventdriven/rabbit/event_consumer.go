package rabbit

import (
	"github.com/ztimes2/jazzba/pkg/eventdriven"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type acknowledger struct {
	channel *amqp.Channel
}

func newAcknowledger(channel *amqp.Channel) *acknowledger {
	return &acknowledger{
		channel: channel,
	}
}

func (a acknowledger) Ack(event eventdriven.Event) error {
	return a.channel.Ack(uint64(event.AcknowledgementTag), false)
}

func (a acknowledger) Nack(event eventdriven.Event) error {
	return a.channel.Nack(uint64(event.AcknowledgementTag), false, true)
}

// EventConsumer implements eventdriven.Consumer interface and provides
// functionality for consuming events from RabbitMQ.
type EventConsumer struct {
	connection *amqp.Connection
	eventTypes []eventdriven.EventType
}

// NewEventConsumer initializes a new instance of EventConsumer.
func NewEventConsumer(conn *amqp.Connection, eventTypes ...eventdriven.EventType,
) *EventConsumer {
	return &EventConsumer{
		connection: conn,
		eventTypes: eventTypes,
	}
}

// consume starts consumption of event messages from a single RabbitMQ's queue
// associated with an event type.
func (ec EventConsumer) consume(eventType eventdriven.EventType,
) (<-chan eventdriven.Delivery, error) {

	queueName, ok := eventQueuesMap[eventType]
	if !ok {
		return nil, newUnknownEventError(eventType)
	}

	channel, err := ec.connection.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "could not initialise amqp channel")
	}

	messageChan, err := channel.Consume(
		// Queue name for consuming.
		queueName,
		// Consumer name is empty because we don't need consumer identification.
		"",
		// Auto-ackowledgement is turned off because an acknowledger is returned
		// by the function and a caller can manually acknowledge deliveries.
		false,
		// Exclusiveness is turned off because it's unnecessary.
		false,
		// NoLocal flag is not supported by RabbitMQ.
		false,
		// FIXME: Not sure what does noWait flag do. Keeping it off for now.
		false,
		// No extra arguments are necessary.
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not start consuming messages from queue")
	}

	acknowledger := newAcknowledger(channel)

	deliveryChan := make(chan eventdriven.Delivery)
	go func() {
		for message := range messageChan {
			deliveryChan <- eventdriven.Delivery{
				Event: eventdriven.Event{
					ID:                 message.MessageId,
					Type:               eventType,
					ContentType:        message.ContentType,
					Body:               message.Body,
					AcknowledgementTag: int(message.DeliveryTag),
				},
				Acknowledger: acknowledger,
			}
		}
	}()
	return deliveryChan, nil
}

// Consume starts consumption of event messages from RabbitMQ.
func (ec EventConsumer) Consume() (<-chan eventdriven.Delivery, error) {
	aggregatedDeliveryChan := make(chan eventdriven.Delivery)

	var deliveryChans []<-chan eventdriven.Delivery
	for _, eventType := range ec.eventTypes {
		deliveryChan, err := ec.consume(eventType)
		if err != nil {
			return nil, errors.Wrapf(
				err, "could not start consumption of events '%s'", eventType)
		}
		deliveryChans = append(deliveryChans, deliveryChan)
	}

	for _, deliveryChan := range deliveryChans {
		go func(deliveryChan <-chan eventdriven.Delivery) {
			for delivery := range deliveryChan {
				aggregatedDeliveryChan <- delivery
			}
		}(deliveryChan)
	}
	return aggregatedDeliveryChan, nil
}
