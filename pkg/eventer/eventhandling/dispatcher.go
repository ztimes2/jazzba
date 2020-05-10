package eventhandling

import (
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DispatcherConfig holds parameters required for initializing an event dispatcher.
type DispatcherConfig struct {
	EventConsumer   eventdriven.Consumer
	NoteUpdater     search.NoteUpdater
	NoteTagUpdater  search.NoteTagUpdater
	NotebookUpdater search.NotebookUpdater
	NoteStore       storage.NoteStore
	NoteTagStore    storage.NoteTagStore
	NotebookStore   storage.NotebookStore
	Logger          logrus.FieldLogger
}

// Dispatcher provides functionality for consuming and handling events continuously.
type Dispatcher struct {
	eventConsumer eventdriven.Consumer
	eventHandlers eventdriven.Handlers
	logger        logrus.FieldLogger
}

// NewDispatcher initializes a new instance of Dispatcher.
func NewDispatcher(cfg DispatcherConfig) *Dispatcher {
	noteEventHandler := newNoteEventHandler(cfg.NoteUpdater, cfg.NoteStore,
		cfg.NoteTagStore, cfg.NotebookStore)
	noteTagEventHandler := newNoteTagEventHandler(cfg.NoteTagUpdater,
		cfg.NoteTagStore)
	notebookEventHandler := newNotebookEventHandler(cfg.NotebookUpdater,
		cfg.NotebookStore)

	eventHandlers := eventdriven.Handlers{}
	eventHandlers.Register(eventdriven.EventTypeNoteCreated, noteEventHandler.noteCreated)
	eventHandlers.Register(eventdriven.EventTypeNoteUpdated, noteEventHandler.noteUpdated)
	eventHandlers.Register(eventdriven.EventTypeNoteDeleted, noteEventHandler.noteDeleted)
	eventHandlers.Register(eventdriven.EventTypeNoteTagCreated, noteTagEventHandler.noteTagCreated)
	eventHandlers.Register(eventdriven.EventTypeNoteTagDeleted, noteTagEventHandler.noteTagDeleted)
	eventHandlers.Register(eventdriven.EventTypeNotebookUpdated, notebookEventHandler.notebookUpdated)
	eventHandlers.Register(eventdriven.EventTypeNotebookDeleted, notebookEventHandler.notebookDeleted)

	return &Dispatcher{
		eventConsumer: cfg.EventConsumer,
		eventHandlers: eventHandlers,
		logger:        cfg.Logger,
	}
}

// ConsumeAndDispatch starts consuming events and dispatching them to registered
// event handlers.
func (d Dispatcher) ConsumeAndDispatch() error {
	deliveryChan, err := d.eventConsumer.Consume()
	if err != nil {
		return errors.Wrap(err, "could not start consuming events")
	}

	for delivery := range deliveryChan {
		handlerFn, ok := d.eventHandlers.Get(delivery.Event.Type)
		if !ok {
			d.logger.Errorf("handler is not registered for event '%s'",
				delivery.Event.Type)
			continue
		}

		if err := handlerFn(delivery.Event); err != nil {
			d.logger.Errorf("could not handle event '%s' with payload '%s': %w",
				delivery.Event.Type, string(delivery.Event.Body), err)
			delivery.Nack()
			continue
		}

		delivery.Ack()
		d.logger.Infof("successfully handled event '%s' with payload '%s'",
			delivery.Event.Type, string(delivery.Event.Body))
	}
	return nil
}
