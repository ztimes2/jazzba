package service

import (
	"github.com/ztimes2/jazzba/pkg/api/p8n"
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/sirupsen/logrus"

	validator "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// CreateNoteParameters holds parameters for creating a note.
type CreateNoteParameters struct {
	Name       string `validate:"required,max=200"`
	Content    string `validate:"required,max=5000"`
	NotebookID int
}

// UpdateNoteParameters holds parameters for updating a note.
type UpdateNoteParameters struct {
	NoteID     int
	Name       string `validate:"required,max=200"`
	Content    string `validate:"required,max=5000"`
	NotebookID int
}

// PaginatedNotes represents a paginated list of notes.
type PaginatedNotes struct {
	Notes      []storage.Note
	Pagination p8n.Pagination
}

// Noter abstracts functionality for manipulating notes.
type Noter interface {
	CreateNote(CreateNoteParameters) (*storage.Note, error)
	FetchNote(noteID int) (*storage.Note, error)
	FetchNotes(p8n.Page) (*PaginatedNotes, error)
	FetchNotesByNotebook(notebookID int, page p8n.Page) (*PaginatedNotes, error)
	FetchNotesByNotebooks(notebookIDs []int) (storage.NotebookNotesMap, error)
	FetchNotesBySearchQuery(query string, page p8n.Page) (*PaginatedNotes, error)
	UpdateNote(UpdateNoteParameters) (*storage.Note, error)
	DeleteNote(noteID int) error
}

// NoteService implements Noter interface and provides functionality for
// manipulating notes.
type NoteService struct {
	noteStore     storage.NoteStore
	noteSearcher  search.NoteSearcher
	validator     *validator.Validate
	eventProducer eventdriven.Producer
	logger        logrus.FieldLogger
}

// NewNoteService initializes a new instance of NoteService.
func NewNoteService(
	noteStore storage.NoteStore,
	noteSearcher search.NoteSearcher,
	validator *validator.Validate,
	eventProducer eventdriven.Producer,
	logger logrus.FieldLogger) *NoteService {
	return &NoteService{
		noteStore:     noteStore,
		noteSearcher:  noteSearcher,
		validator:     validator,
		eventProducer: eventProducer,
		logger:        logger,
	}
}

func newNoteStoreStartTransactionError(err error) error {
	return errors.Wrap(err, "could not start a note store transaction")
}

func newGeneralNoteStoreError(err error) error {
	return errors.Wrap(err, "note store failure")
}

// CreateNote creates a new note.
func (ns NoteService) CreateNote(params CreateNoteParameters) (*storage.Note, error) {
	if err := ns.validator.Struct(&params); err != nil {
		return nil, err
	}

	tx, err := ns.noteStore.BeginTx()
	if err != nil {
		return nil, newNoteStoreStartTransactionError(err)
	}

	createdNote, err := ns.noteStore.CreateOne(storage.CreateNoteParameters{
		Transaction: tx,
		Name:        params.Name,
		Content:     params.Content,
		NotebookID:  params.NotebookID,
	})
	if err != nil {
		tx.Rollback()
		return nil, newGeneralNoteStoreError(err)
	}
	tx.Commit()

	eventType := eventdriven.EventTypeNoteCreated
	if err := ns.eventProducer.Produce(eventType,
		eventdriven.NoteCreatedEventPayload{
			NoteID: createdNote.ID,
		},
	); err != nil {
		ns.logger.Error(newEventProducingError(eventType, err))
	}

	return createdNote, nil
}

// FetchNote fetches a note.
func (ns NoteService) FetchNote(noteID int) (*storage.Note, error) {
	note, err := ns.noteStore.FetchOne(noteID)
	if err != nil {
		return nil, newGeneralNoteStoreError(err)
	}
	return note, nil
}

// FetchNotes fetches all notes and paginates them.
func (ns NoteService) FetchNotes(page p8n.Page) (*PaginatedNotes, error) {
	notes, err := ns.noteStore.FetchAllPaginated(page.Limit, page.Offset)
	if err != nil {
		return nil, newGeneralNoteStoreError(err)
	}
	return &PaginatedNotes{
		Notes:      notes,
		Pagination: p8n.NewPagination(len(notes), page),
	}, nil
}

// FetchNotesByNotebook fetches all notes associated with a notebook and
// paginates them.
func (ns NoteService) FetchNotesByNotebook(notebookID int, page p8n.Page,
) (*PaginatedNotes, error) {

	notes, err := ns.noteStore.FetchManyByNotebookPaginated(
		notebookID, page.Limit, page.Offset)
	if err != nil {
		return nil, newGeneralNoteStoreError(err)
	}
	return &PaginatedNotes{
		Notes:      notes,
		Pagination: p8n.NewPagination(len(notes), page),
	}, nil
}

// FetchNotesByNotebooks fetches all notes associated with notebooks.
func (ns NoteService) FetchNotesByNotebooks(notebookIDs []int,
) (storage.NotebookNotesMap, error) {

	notebookNotesMap, err := ns.noteStore.FetchManyByNotebooks(notebookIDs)
	if err != nil {
		// TODO: check error cases
		return nil, newGeneralNoteStoreError(err)
	}
	return notebookNotesMap, nil
}

// FetchNotesBySearchQuery fetches notes by performing a full-text search.
func (ns NoteService) FetchNotesBySearchQuery(query string, page p8n.Page,
) (*PaginatedNotes, error) {

	searchNotes, err := ns.noteSearcher.SearchByQuery(
		query, page.Limit, page.Offset)
	if err != nil {
		return nil, errors.Wrap(err, "note searcher failure")
	}

	if len(searchNotes) == 0 {
		return &PaginatedNotes{
			Pagination: p8n.NewPagination(0, page),
		}, nil
	}

	var noteIDs []int
	for _, note := range searchNotes {
		noteIDs = append(noteIDs, note.ID)
	}

	storageNotes, err := ns.noteStore.FetchMany(noteIDs)
	if err != nil {
		// TODO: check error cases
		return nil, newGeneralNoteStoreError(err)
	}

	return &PaginatedNotes{
		Notes:      storageNotes,
		Pagination: p8n.NewPagination(len(storageNotes), page),
	}, nil
}

// UpdateNote updates a note.
func (ns NoteService) UpdateNote(params UpdateNoteParameters) (*storage.Note, error) {
	if err := ns.validator.Struct(&params); err != nil {
		return nil, err
	}

	tx, err := ns.noteStore.BeginTx()
	if err != nil {
		return nil, newNoteStoreStartTransactionError(err)
	}

	updatedNote, err := ns.noteStore.UpdateOne(storage.UpdateNoteParameters{
		Transaction: tx,
		NoteID:      params.NoteID,
		Name:        params.Name,
		Content:     params.Content,
		NotebookID:  params.NotebookID,
	})
	if err != nil {
		tx.Rollback()
		return nil, newGeneralNoteStoreError(err)
	}
	tx.Commit()

	eventType := eventdriven.EventTypeNoteUpdated
	if err := ns.eventProducer.Produce(eventType,
		eventdriven.NoteUpdatedEventPayload{
			NoteID: updatedNote.ID,
		},
	); err != nil {
		ns.logger.Error(newEventProducingError(eventType, err))
	}

	return updatedNote, nil
}

// DeleteNote deletes a note.
func (ns NoteService) DeleteNote(noteID int) error {
	tx, err := ns.noteStore.BeginTx()
	if err != nil {
		return newNoteStoreStartTransactionError(err)
	}

	if err := ns.noteStore.DeleteOne(tx, noteID); err != nil {
		tx.Rollback()
		return newGeneralNoteStoreError(err)
	}
	tx.Commit()

	eventType := eventdriven.EventTypeNoteDeleted
	if err := ns.eventProducer.Produce(eventType,
		eventdriven.NoteDeletedEventPayload{
			NoteID: noteID,
		},
	); err != nil {
		ns.logger.Error(newEventProducingError(eventType, err))
	}

	return nil
}
