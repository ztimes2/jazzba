package rabbit

import (
	"github.com/ztimes2/jazzba/pkg/eventdriven"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

const (
	queueNameNoteCreated = "notes.created"
	queueNameNoteUpdated = "notes.updated"
	queueNameNoteDeleted = "notes.deleted"

	queueNameNoteTagCreated = "note_tags.created"
	queueNameNoteTagDeleted = "note_tags.deleted"

	queueNameNotebookUpdated = "notebooks.updated"
	queueNameNotebookDeleted = "notebooks.deleted"
)

// Holds mapping of event types and RabbitMQ's queue names associated with them.
var eventQueuesMap = map[eventdriven.EventType]string{
	eventdriven.EventTypeNoteCreated: queueNameNoteCreated,
	eventdriven.EventTypeNoteUpdated: queueNameNoteUpdated,
	eventdriven.EventTypeNoteDeleted: queueNameNoteDeleted,

	eventdriven.EventTypeNoteTagCreated: queueNameNoteTagCreated,
	eventdriven.EventTypeNoteTagDeleted: queueNameNoteTagDeleted,

	eventdriven.EventTypeNotebookUpdated: queueNameNotebookUpdated,
	eventdriven.EventTypeNotebookDeleted: queueNameNotebookDeleted,
}

const (
	contentTypeJSON = "application/json"
)

// Config holds configuration for connecting to RabbitMQ.
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

func (c Config) toURL() string {
	return "amqp://" + c.Username + ":" + c.Password + "@" + c.Host + ":" + c.Port
}

// NewConnection initializes a new RabbitMQ connection.
func NewConnection(cfg Config) (*amqp.Connection, error) {
	return amqp.Dial(cfg.toURL())
}

// DeclareQueues creates all necessary queues in RabbitMQ if they do not exist.
func DeclareQueues(conn *amqp.Connection) error {
	channel, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "could not initialise amqp channel")
	}
	defer channel.Close()

	queues := []string{
		queueNameNoteCreated,
		queueNameNoteUpdated,
		queueNameNoteDeleted,

		queueNameNoteTagCreated,
		queueNameNoteTagDeleted,

		queueNameNotebookUpdated,
		queueNameNotebookDeleted,
	}

	for _, queue := range queues {
		if _, err := channel.QueueDeclare(
			queue,
			false,
			false,
			false,
			false,
			nil,
		); err != nil {
			return errors.Wrapf(err, "could not declare queue '%s'", queue)
		}
	}
	return nil
}

func newUnknownEventError(eventType eventdriven.EventType) error {
	return errors.Errorf("unknown event '%s'", eventType)
}
