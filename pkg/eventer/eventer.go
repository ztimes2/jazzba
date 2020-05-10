package eventer

import (
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/eventer/eventhandling"
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/sirupsen/logrus"
)

// Dependencies holds dependencies and primitives of the eventer.
type Dependencies struct {
	EventConsumer   eventdriven.Consumer
	NoteStore       storage.NoteStore
	NoteTagStore    storage.NoteTagStore
	NotebookStore   storage.NotebookStore
	NoteUpdater     search.NoteUpdater
	NoteTagUpdater  search.NoteTagUpdater
	NotebookUpdater search.NotebookUpdater
	Logger          logrus.FieldLogger
}

// Eventer provides functionality for updating components of the service based
// on different events.
type Eventer struct {
	logger          logrus.FieldLogger
	eventDispatcher *eventhandling.Dispatcher
}

// New initializes a new instance of Eventer.
func New(deps Dependencies) *Eventer {
	return &Eventer{
		eventDispatcher: eventhandling.NewDispatcher(eventhandling.DispatcherConfig{
			EventConsumer:   deps.EventConsumer,
			NoteUpdater:     deps.NoteUpdater,
			NoteTagUpdater:  deps.NoteTagUpdater,
			NotebookUpdater: deps.NotebookUpdater,
			NoteStore:       deps.NoteStore,
			NoteTagStore:    deps.NoteTagStore,
			NotebookStore:   deps.NotebookStore,
			Logger:          deps.Logger,
		}),
		logger: deps.Logger,
	}
}

// Run starts consuming and handling events.
func (e Eventer) Run() {
	// TODO: implement a graceful shutdown

	e.logger.Info("eventer is started")
	defer e.logger.Info("eventer is stopped")

	if err := e.eventDispatcher.ConsumeAndDispatch(); err != nil {
		e.logger.Error(err)
	}
}
