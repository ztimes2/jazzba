package eventhandling

import (
	"encoding/json"

	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/pkg/errors"
)

type noteEventHandler struct {
	noteUpdater   search.NoteUpdater
	noteStore     storage.NoteStore
	noteTagStore  storage.NoteTagStore
	notebookStore storage.NotebookStore
}

func newNoteEventHandler(noteUpdater search.NoteUpdater,
	noteStore storage.NoteStore, noteTagStore storage.NoteTagStore,
	notebookStore storage.NotebookStore) *noteEventHandler {
	return &noteEventHandler{
		noteUpdater:   noteUpdater,
		noteStore:     noteStore,
		noteTagStore:  noteTagStore,
		notebookStore: notebookStore,
	}
}

func (neh noteEventHandler) noteCreated(event eventdriven.Event) error {
	var payload eventdriven.NoteCreatedEventPayload
	if err := json.Unmarshal(event.Body, &payload); err != nil {
		return newPayloadDecodingError(err)
	}

	note, err := neh.noteStore.FetchOne(payload.NoteID)
	if err != nil {
		return errors.Wrapf(err, "could not fetch note(%d) from data store",
			payload.NoteID)
	}

	notebook, err := neh.notebookStore.FetchOne(note.NotebookID)
	if err != nil {
		return errors.Wrapf(err,
			"could not fetch notebook(%d) associated with note(%d) from data store",
			note.NotebookID, note.ID)
	}

	noteTags, err := neh.noteTagStore.FetchManyByNote(note.ID)
	if err != nil {
		return errors.Wrapf(err,
			"could not fetch tags associated with note(%d) from data store",
			note.ID)
	}

	if err := neh.noteUpdater.CreateOne(
		newNote(*note, *notebook, noteTags),
	); err != nil {
		var dErr *search.DuplicateError
		if errors.As(err, &dErr) {
			return nil
		}

		return errors.Wrapf(
			err, "could not create note(%d) in search engine", note.ID)
	}
	return nil
}

func (neh noteEventHandler) noteUpdated(event eventdriven.Event) error {
	var payload eventdriven.NoteUpdatedEventPayload
	if err := json.Unmarshal(event.Body, &payload); err != nil {
		return newPayloadDecodingError(err)
	}

	note, err := neh.noteStore.FetchOne(payload.NoteID)
	if err != nil {
		return errors.Wrapf(
			err, "could not fetch note(%d) from note store", payload.NoteID)
	}

	notebook, err := neh.notebookStore.FetchOne(note.NotebookID)
	if err != nil {
		return errors.Wrapf(
			err,
			"could not fetch notebook(%d) associated with note(%d) from notebook store",
			note.NotebookID, note.ID)
	}

	noteTags, err := neh.noteTagStore.FetchManyByNote(note.ID)
	if err != nil {
		return errors.Wrapf(
			err,
			"could not fetch tags belonging to note(%d) from note tag store",
			note.ID)
	}

	if err := neh.noteUpdater.UpdateOne(
		newNote(*note, *notebook, noteTags),
	); err != nil {
		var nfErr *search.NotFoundError
		if errors.As(err, &nfErr) {
			return nil
		}

		return errors.Wrapf(
			err, "could not update note(%d) in search engine", note.ID)
	}
	return nil
}

func (neh noteEventHandler) noteDeleted(event eventdriven.Event) error {
	var payload eventdriven.NoteDeletedEventPayload
	if err := json.Unmarshal(event.Body, &payload); err != nil {
		return newPayloadDecodingError(err)
	}

	if err := neh.noteUpdater.DeleteOne(payload.NoteID); err != nil {
		var nfErr *search.NotFoundError
		if errors.As(err, &nfErr) {
			return nil
		}

		return errors.Wrapf(err,
			"could not delete note(%d) from search engine", payload.NoteID)
	}
	return nil
}
