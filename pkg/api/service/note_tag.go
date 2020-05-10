package service

import (
	"github.com/ztimes2/jazzba/pkg/api/p8n"
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/storage"

	validator "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CreateNoteTagParameters holds parameters for creating a note tag.
type CreateNoteTagParameters struct {
	NoteID  int
	TagName string `validate:"required,max=100"`
}

// PaginatedNoteTags represents a paginated list of note tags.
type PaginatedNoteTags struct {
	NoteTags   []storage.NoteTag
	Pagination p8n.Pagination
}

// NoteTagger abstracts functionality for manipulating note tags.
type NoteTagger interface {
	CreateNoteTag(CreateNoteTagParameters) (*storage.NoteTag, error)
	FetchNoteTagsByNote(noteID int, page p8n.Page) (*PaginatedNoteTags, error)
	FetchNoteTagsByNotes(noteIDs []int) (storage.NoteTagsMap, error)
	DeleteNoteTag(noteID int, tagName string) error
}

// NoteTagService implements NoteTagger interface and provides functionality for
// manipulating note tags.
type NoteTagService struct {
	noteTagStore  storage.NoteTagStore
	validator     *validator.Validate
	eventProducer eventdriven.Producer
	logger        logrus.FieldLogger
}

// NewNoteTagService initializes a new instance of NoteTagService.
func NewNoteTagService(
	noteTagStore storage.NoteTagStore,
	validator *validator.Validate,
	eventProducer eventdriven.Producer,
	logger logrus.FieldLogger) *NoteTagService {
	return &NoteTagService{
		noteTagStore:  noteTagStore,
		validator:     validator,
		eventProducer: eventProducer,
		logger:        logger,
	}
}

func newNoteTagStoreStartTransactionError(err error) error {
	return errors.Wrap(err, "could not start a note tag store transaction")
}

func newGeneralNoteTagStoreError(err error) error {
	return errors.Wrap(err, "note tag store failure")
}

// CreateNoteTag adds a new tag associated with a note.
func (nts NoteTagService) CreateNoteTag(
	params CreateNoteTagParameters) (*storage.NoteTag, error) {

	if err := nts.validator.Struct(&params); err != nil {
		return nil, err
	}

	tx, err := nts.noteTagStore.BeginTx()
	if err != nil {
		return nil, newNoteTagStoreStartTransactionError(err)
	}

	createdNoteTag, err :=
		nts.noteTagStore.CreateOne(tx, storage.CreateNoteTagParameters{
			NoteID:  params.NoteID,
			TagName: params.TagName,
		})
	if err != nil {
		tx.Rollback()
		return nil, newGeneralNoteTagStoreError(err)
	}
	tx.Commit()

	eventType := eventdriven.EventTypeNoteTagCreated
	if err := nts.eventProducer.Produce(eventType,
		eventdriven.NoteTagCreatedEventPayload{
			NoteID: createdNoteTag.NoteID,
		},
	); err != nil {
		nts.logger.Error(newEventProducingError(eventType, err))
	}

	return createdNoteTag, nil
}

// FetchNoteTagsByNote fetches all tags associated with a note and paginates them.
func (nts NoteTagService) FetchNoteTagsByNote(noteID int, page p8n.Page,
) (*PaginatedNoteTags, error) {

	noteTags, err := nts.noteTagStore.FetchManyByNotePaginated(
		noteID, page.Limit, page.Offset)
	if err != nil {
		return nil, newGeneralNoteTagStoreError(err)
	}

	return &PaginatedNoteTags{
		NoteTags:   noteTags,
		Pagination: p8n.NewPagination(len(noteTags), page),
	}, nil
}

// FetchNoteTagsByNotes fetches all tags associated with notes.
func (nts NoteTagService) FetchNoteTagsByNotes(
	noteIDs []int) (storage.NoteTagsMap, error) {

	noteTagsMap, err := nts.noteTagStore.FetchManyByNotes(noteIDs)
	if err != nil {
		return nil, newGeneralNoteTagStoreError(err)
	}
	return noteTagsMap, nil
}

// DeleteNoteTag deletes a tag associated with a note.
func (nts NoteTagService) DeleteNoteTag(noteID int, tagName string) error {
	tx, err := nts.noteTagStore.BeginTx()
	if err != nil {
		return newNoteTagStoreStartTransactionError(err)
	}

	if err := nts.noteTagStore.DeleteOne(tx, storage.DeleteNoteTagParameters{
		NoteID:  noteID,
		TagName: tagName,
	}); err != nil {
		tx.Rollback()
		return newGeneralNoteTagStoreError(err)
	}
	tx.Commit()

	eventType := eventdriven.EventTypeNoteTagDeleted
	if err := nts.eventProducer.Produce(eventType,
		eventdriven.NoteTagDeletedEventPayload{
			NoteID: noteID,
		},
	); err != nil {
		nts.logger.Error(newEventProducingError(eventType, err))
	}

	return nil
}
