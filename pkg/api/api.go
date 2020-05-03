package api

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/ztimes2/jazzba/pkg/api/httphandling"
	"github.com/ztimes2/jazzba/pkg/api/service"
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"
)

// Dependencies holds dependencies and primitives of the service's HTTP API.
type Dependencies struct {
	Logger        logrus.FieldLogger
	NotebookStore storage.NotebookStore
	NoteStore     storage.NoteStore
	NoteTagStore  storage.NoteTagStore
	NoteSearcher  search.NoteSearcher
	EventProducer eventdriven.Producer
}

// API represents an HTTP API of the service.
type API struct {
	logger logrus.FieldLogger
	server *http.Server
}

// New initializes a new instance of API.
func New(serverPort string, deps Dependencies) *API {
	validator := validator.New()

	router := httphandling.NewRouter(httphandling.RouterConfig{
		NotebookService: service.NewNotebookService(
			deps.NotebookStore, validator, deps.EventProducer, deps.Logger,
		),
		NoteService: service.NewNoteService(
			deps.NoteStore, deps.NoteSearcher, validator, deps.EventProducer,
			deps.Logger,
		),
		NoteTagService: service.NewNoteTagService(
			deps.NoteTagStore, validator, deps.EventProducer, deps.Logger,
		),
		Logger: deps.Logger,
	})

	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: router,
		// TODO: set time outs
	}

	return &API{
		server: server,
		logger: deps.Logger,
	}
}

// Run spins up an HTTP server and starts accepting HTTP requests.
func (a API) Run() {
	// TODO: implement a graceful shutdown

	a.logger.Info("server started on " + a.server.Addr)
	defer a.logger.Info("server stopped")

	if err := a.server.ListenAndServe(); err != nil {
		a.logger.Error(err)
	}
}
