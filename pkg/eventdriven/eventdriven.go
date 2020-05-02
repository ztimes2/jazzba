package eventdriven

// EventType represents a type of an event.
type EventType string

const (
	// EventTypeNoteCreated represents an event when a new note is created in a
	// data store.
	EventTypeNoteCreated EventType = "note_created"

	// EventTypeNoteUpdated represents an event when an existing note is updated
	// in a data store.
	EventTypeNoteUpdated EventType = "note_updated"

	// EventTypeNoteDeleted represents an event when an existing note is deleted
	// from a data store.
	EventTypeNoteDeleted EventType = "note_deleted"

	// EventTypeNoteTagCreated represents an event when a new tag is added to an
	// existing note in a data store.
	EventTypeNoteTagCreated EventType = "note_tag_created"

	// EventTypeNoteTagDeleted represents an event when an existing tag is
	// removed from an existing note in a data store.
	EventTypeNoteTagDeleted EventType = "note_tag_deleted"

	// EventTypeNotebookUpdated represents an event when an existing notebook
	// is updated in a date store.
	EventTypeNotebookUpdated EventType = "notebook_updated"

	// EventTypeNotebookDeleted represents an event when an existing notebook
	// is deleted from a data store.
	EventTypeNotebookDeleted EventType = "notebook_deleted"
)

// Event represents an event.
type Event struct {
	ID                 string
	Type               EventType
	AcknowledgementTag int
	ContentType        string
	Body               []byte
}

// Producer defines functionality for producing events to an event queue.
type Producer interface {
	Produce(eventType EventType, payload interface{}) error
}

// Acknowledger abstracts functionality for acknowledging events received from
// an event queue.
type Acknowledger interface {
	Ack(Event) error
	Nack(Event) error
}

// Delivery holds an event and its acknowledger received from an event queue.
type Delivery struct {
	Event        Event
	Acknowledger Acknowledger
}

// Ack acknowledges a delivery's event as successfully handled.
func (d Delivery) Ack() error {
	return d.Acknowledger.Ack(d.Event)
}

// Nack acknowledges a delivery's event as unsuccessfully handled.
func (d Delivery) Nack() error {
	return d.Acknowledger.Nack(d.Event)
}

// Consumer abstracts functionality for consuming events from an event queue.
type Consumer interface {
	Consume() (<-chan Delivery, error)
}

// HandlerFunc represents a function which handles an event.
type HandlerFunc func(Event) error

// Handlers represents a mapping of events and handlers registered for handling
// them.
type Handlers map[EventType]HandlerFunc

// Register registers a handler for an event.
func (h Handlers) Register(eventType EventType, handlerFn HandlerFunc) {
	if h == nil {
		h = make(map[EventType]HandlerFunc)
	}
	h[eventType] = handlerFn
}

// Get returns a handler for an event if it's registered.
func (h Handlers) Get(eventType EventType) (HandlerFunc, bool) {
	if h == nil {
		return nil, false
	}

	handlerFn, ok := h[eventType]
	if !ok {
		return nil, false
	}

	return handlerFn, true
}
