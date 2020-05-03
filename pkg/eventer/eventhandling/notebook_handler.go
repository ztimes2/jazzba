package eventhandling

import (
	"encoding/json"

	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/pkg/errors"
)

type notebookEventHandler struct {
	notebookUpdater search.NotebookUpdater
	notebookStore   storage.NotebookStore
}

func newNotebookEventHandler(notebookUpdater search.NotebookUpdater,
	notebookStore storage.NotebookStore) *notebookEventHandler {
	return &notebookEventHandler{
		notebookUpdater: notebookUpdater,
		notebookStore:   notebookStore,
	}
}

func (neh notebookEventHandler) notebookUpdated(event eventdriven.Event) error {
	var payload eventdriven.NotebookUpdatedEventPayload
	if err := json.Unmarshal(event.Body, &payload); err != nil {
		return newPayloadDecodingError(err)
	}

	notebook, err := neh.notebookStore.FetchOne(payload.NotebookID)
	if err != nil {
		return errors.Wrapf(err, "could not fetch notebook(%d) from data store",
			payload.NotebookID)
	}

	if err := neh.notebookUpdater.UpdateOne(search.Notebook{
		ID:   notebook.ID,
		Name: notebook.Name,
	}); err != nil {
		var nfErr *search.NotFoundError
		if errors.As(err, &nfErr) {
			return nil
		}

		return errors.Wrapf(err, "could not update notebook(%d) in search engine",
			notebook.ID)
	}
	return nil
}

func (neh notebookEventHandler) notebookDeleted(event eventdriven.Event) error {
	var payload eventdriven.NotebookDeletedEventPayload
	if err := json.Unmarshal(event.Body, &payload); err != nil {
		return newPayloadDecodingError(err)
	}

	if err := neh.notebookUpdater.DeleteOne(payload.NotebookID); err != nil {
		var nfErr *search.NotFoundError
		if errors.As(err, &nfErr) {
			return nil
		}

		return errors.Wrapf(err, "could not delete notebook(%d) from search engine",
			payload.NotebookID)
	}
	return nil
}
