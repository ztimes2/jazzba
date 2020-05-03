package eventhandling

import (
	"encoding/json"

	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/pkg/errors"
)

type noteTagEventHandler struct {
	noteTagUpdater search.NoteTagUpdater
	noteTagStore   storage.NoteTagStore
}

func newNoteTagEventHandler(noteTagUpdater search.NoteTagUpdater,
	noteTagStore storage.NoteTagStore) *noteTagEventHandler {
	return &noteTagEventHandler{
		noteTagUpdater: noteTagUpdater,
		noteTagStore:   noteTagStore,
	}
}

func (nteh noteTagEventHandler) noteTagCreated(event eventdriven.Event) error {
	var payload eventdriven.NoteTagCreatedEventPayload
	if err := json.Unmarshal(event.Body, &payload); err != nil {
		return newPayloadDecodingError(err)
	}

	noteTags, err := nteh.noteTagStore.FetchManyByNote(payload.NoteID)
	if err != nil {
		return errors.Wrapf(err,
			"could not fetch tags associated with note(%d) from data store",
			payload.NoteID)
	}

	if err := nteh.noteTagUpdater.UpdateMany(
		payload.NoteID, extractTagsFromNoteTags(noteTags),
	); err != nil {
		var nfErr *search.NotFoundError
		if errors.As(err, &nfErr) {
			return nil
		}

		return errors.Wrapf(err,
			"could not update tags associated with note(%d) in search engine",
			payload.NoteID)
	}
	return nil
}

func (nteh noteTagEventHandler) noteTagDeleted(event eventdriven.Event) error {
	var payload eventdriven.NoteTagDeletedEventPayload
	if err := json.Unmarshal(event.Body, &payload); err != nil {
		return newPayloadDecodingError(err)
	}

	noteTags, err := nteh.noteTagStore.FetchManyByNote(payload.NoteID)
	if err != nil {
		return errors.Wrapf(err,
			"could not fetch tags associated with note(%d) from data store",
			payload.NoteID)
	}

	if err := nteh.noteTagUpdater.UpdateMany(
		payload.NoteID, extractTagsFromNoteTags(noteTags),
	); err != nil {
		var nfErr *search.NotFoundError
		if errors.As(err, &nfErr) {
			return nil
		}

		return errors.Wrapf(err,
			"could not update tags associated with note(%d) in search engine",
			payload.NoteID)
	}
	return nil
}
