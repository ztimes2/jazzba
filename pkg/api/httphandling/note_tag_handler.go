package httphandling

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ztimes2/jazzba/pkg/api/service"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/sirupsen/logrus"
)

type noteTagHandler struct {
	noteTagService service.NoteTagger
	logger         logrus.FieldLogger
}

func newNoteTagHandler(noteTagService service.NoteTagger, logger logrus.FieldLogger,
) *noteTagHandler {
	return &noteTagHandler{
		noteTagService: noteTagService,
		logger:         logger,
	}
}

func (nth noteTagHandler) createNoteTag(w http.ResponseWriter, r *http.Request) {
	noteID, _ := readIntPathParam(r, pathParamNoteID)

	var reqBody createNoteTagRequestBody
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeBadRequest(w, r.Header, newErrorResponse(err.Error()))
		return
	}

	noteTag, err := nth.noteTagService.CreateNoteTag(
		service.CreateNoteTagParameters{
			NoteID:  noteID,
			TagName: reqBody.TagName,
		},
	)
	if err != nil {
		var drErr *storage.DuplicateError
		if errors.As(err, &drErr) {
			writeBadRequest(w, r.Header, newErrorResponse(drErr.Error()))
			return
		}

		var isrrErr *storage.ReferenceError
		if errors.As(err, &isrrErr) {
			writeNotFound(w, r.Header, newErrorResponse(isrrErr.Error()))
			return
		}

		writeInternalServerError(w, nth.logger, err)
		return
	}

	writeCreated(w, r.Header, newNoteTagResponse(*noteTag))
}

func (nth noteTagHandler) fetchNoteTagsByNote(w http.ResponseWriter, r *http.Request) {
	noteID, _ := readIntPathParam(r, pathParamNoteID)
	page := readPaginationParam(r)

	paginatedNoteTags, err := nth.noteTagService.FetchNoteTagsByNote(noteID, page)
	if err != nil {
		// TODO: handle case when the given note does not exist
		writeInternalServerError(w, nth.logger, err)
		return
	}

	writeOK(w, r.Header, newPaginatedNoteTagsResponse(*paginatedNoteTags))
}

func (nth noteTagHandler) fetchNoteTagsByNotes(w http.ResponseWriter, r *http.Request) {
	noteIDs, ok := readIntQueryParams(r, "note_id")
	if !ok {
		writeBadRequest(w, r.Header, newErrorResponse("invalid note ID parameters"))
		return
	}

	noteTagsMap, err := nth.noteTagService.FetchNoteTagsByNotes(noteIDs)
	if err != nil {
		writeInternalServerError(w, nth.logger, err)
		return
	}

	writeOK(w, r.Header, newNoteTagsResponse(noteTagsMap))
}

func (nth noteTagHandler) deleteNoteTag(w http.ResponseWriter, r *http.Request) {
	noteID, _ := readIntPathParam(r, pathParamNoteID)
	tagName := readStringPathParam(r, pathParamTagName)

	if err := nth.noteTagService.DeleteNoteTag(noteID, tagName); err != nil {
		var rnfErr *storage.NotFoundError
		if errors.As(err, &rnfErr) {
			writeNotFound(w, r.Header, newErrorResponse(rnfErr.Error()))
		}

		writeInternalServerError(w, nth.logger, err)
		return
	}

	writeNoContent(w)
}
