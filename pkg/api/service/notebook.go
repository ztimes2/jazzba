package service

import (
	"github.com/ztimes2/jazzba/pkg/api/p8n"
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/storage"

	validator "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CreateNotebookParameters holds parameters for creating a notebook.
type CreateNotebookParameters struct {
	Name string `validate:"required,max=200"`
}

// UpdateNotebookParameters holds parameters for updating a notebook.
type UpdateNotebookParameters struct {
	NotebookID int
	Name       string `validate:"required,max=200"`
}

// PaginatedNotebooks represents a paginated list of notebooks.
type PaginatedNotebooks struct {
	Notebooks  []storage.Notebook
	Pagination p8n.Pagination
}

// Notebooker abstracts functionality for manipulating notebooks.
type Notebooker interface {
	CreateNotebook(CreateNotebookParameters) (*storage.Notebook, error)
	FetchNotebook(notebookID int) (*storage.Notebook, error)
	FetchNotebooks(p8n.Page) (*PaginatedNotebooks, error)
	UpdateNotebook(UpdateNotebookParameters) (*storage.Notebook, error)
	DeleteNotebook(notebookID int) error
}

// NotebookService implements Notebooker interface and provides functionality
// for manipulating notebooks.
type NotebookService struct {
	notebookStore storage.NotebookStore
	validator     *validator.Validate
	eventProducer eventdriven.Producer
	logger        logrus.FieldLogger
}

// NewNotebookService initializes a new instance of NotebookService.
func NewNotebookService(
	notebookStore storage.NotebookStore,
	validator *validator.Validate,
	eventProducer eventdriven.Producer,
	logger logrus.FieldLogger) *NotebookService {
	return &NotebookService{
		notebookStore: notebookStore,
		validator:     validator,
		eventProducer: eventProducer,
		logger:        logger,
	}
}

func newNotebookStoreStartTransactionError(err error) error {
	return errors.Wrap(err, "could not start a notebook store transaction")
}

func newGeneralNotebookStoreError(err error) error {
	return errors.Wrap(err, "notebook store failure")
}

// CreateNotebook creates a new notebook.
func (ns NotebookService) CreateNotebook(
	params CreateNotebookParameters) (*storage.Notebook, error) {

	if err := ns.validator.Struct(&params); err != nil {
		return nil, err
	}

	tx, err := ns.notebookStore.BeginTx()
	if err != nil {
		return nil, newNotebookStoreStartTransactionError(err)
	}

	createdNotebook, err := ns.notebookStore.CreateOne(tx, params.Name)
	if err != nil {
		tx.Rollback()
		return nil, newGeneralNotebookStoreError(err)
	}
	tx.Commit()

	return createdNotebook, nil
}

// FetchNotebook fetches a notebook.
func (ns NotebookService) FetchNotebook(notebookID int) (*storage.Notebook, error) {
	notebook, err := ns.notebookStore.FetchOne(notebookID)
	if err != nil {
		return nil, newGeneralNotebookStoreError(err)
	}
	return notebook, nil
}

// FetchNotebooks fetches all notebooks and paginates them.
func (ns NotebookService) FetchNotebooks(page p8n.Page) (*PaginatedNotebooks, error) {
	notebooks, err := ns.notebookStore.FetchAllPaginated(page.Limit, page.Offset)
	if err != nil {
		return nil, newGeneralNotebookStoreError(err)
	}

	return &PaginatedNotebooks{
		Notebooks:  notebooks,
		Pagination: p8n.NewPagination(len(notebooks), page),
	}, nil
}

// UpdateNotebook updates a notebook.
func (ns NotebookService) UpdateNotebook(
	params UpdateNotebookParameters) (*storage.Notebook, error) {

	if err := ns.validator.Struct(&params); err != nil {
		return nil, err
	}

	tx, err := ns.notebookStore.BeginTx()
	if err != nil {
		return nil, newNotebookStoreStartTransactionError(err)
	}

	updatedNotebook, err :=
		ns.notebookStore.UpdateOne(tx, storage.UpdateNotebookParameters{
			NotebookID: params.NotebookID,
			Name:       params.Name,
		})
	if err != nil {
		tx.Rollback()
		return nil, newGeneralNotebookStoreError(err)
	}
	tx.Commit()

	eventType := eventdriven.EventTypeNotebookUpdated
	if err := ns.eventProducer.Produce(eventType,
		eventdriven.NotebookUpdatedEventPayload{
			NotebookID: updatedNotebook.ID,
		},
	); err != nil {
		ns.logger.Error(newEventProducingError(eventType, err))
	}

	return updatedNotebook, nil
}

// DeleteNotebook deletes a notebook.
func (ns NotebookService) DeleteNotebook(notebookID int) error {
	tx, err := ns.notebookStore.BeginTx()
	if err != nil {
		return newNotebookStoreStartTransactionError(err)
	}

	if err := ns.notebookStore.DeleteOne(tx, notebookID); err != nil {
		tx.Rollback()
		return newGeneralNotebookStoreError(err)
	}
	tx.Commit()

	eventType := eventdriven.EventTypeNotebookDeleted
	if err := ns.eventProducer.Produce(eventType,
		eventdriven.NotebookDeletedEventPayload{
			NotebookID: notebookID,
		},
	); err != nil {
		ns.logger.Error(newEventProducingError(eventType, err))
	}

	return nil
}
