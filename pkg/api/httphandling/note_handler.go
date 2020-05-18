package httphandling

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ztimes2/jazzba/pkg/api/service"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/sirupsen/logrus"
)

type noteHandler struct {
	noteService service.Noter
	logger      logrus.FieldLogger
}

func newNoteHandler(noteService service.Noter, logger logrus.FieldLogger,
) *noteHandler {
	return &noteHandler{
		noteService: noteService,
	}
}

func (nh noteHandler) createNote(w http.ResponseWriter, r *http.Request) {
	var reqBody createNoteRequestBody
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeBadRequest(w, r.Header, newErrorResponse(err.Error()))
		return
	}

	note, err := nh.noteService.CreateNote(service.CreateNoteParameters{
		Name:       reqBody.Name,
		Content:    reqBody.Content,
		NotebookID: reqBody.NotebookID,
	})
	if err != nil {
		var dErr *storage.DuplicateError
		if errors.As(err, &dErr) {
			writeBadRequest(w, r.Header, newErrorResponse(dErr.Error()))
			return
		}

		var rErr *storage.ReferenceError
		if errors.As(err, &rErr) {
			writeBadRequest(w, r.Header, newErrorResponse(rErr.Error()))
			return
		}

		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeCreated(w, r.Header, newNoteResponse(*note))
}

func (nh noteHandler) fetchNote(w http.ResponseWriter, r *http.Request) {
	noteID, _ := readIntPathParam(r, pathParamNoteID)

	note, err := nh.noteService.FetchNote(noteID)
	if err != nil {
		var nfErr *storage.NotFoundError
		if errors.As(err, &nfErr) {
			writeBadRequest(w, r.Header, newErrorResponse(nfErr.Error()))
			return
		}

		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeOK(w, r.Header, newNoteResponse(*note))
}

func (nh noteHandler) fetchNotes(w http.ResponseWriter, r *http.Request) {
	page := readPaginationParam(r)

	paginatedNotes, err := nh.noteService.FetchNotes(page)
	if err != nil {
		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeOK(w, r.Header, newPaginatedNotesResponse(*paginatedNotes))
}

func (nh noteHandler) fetchNotesByNotebook(w http.ResponseWriter, r *http.Request) {
	notebookID, _ := readIntPathParam(r, pathParamNotebookID)
	page := readPaginationParam(r)

	paginatedNotes, err := nh.noteService.FetchNotesByNotebook(notebookID, page)
	if err != nil {
		// TODO: handle case when the given notebook does not exist
		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeOK(w, r.Header, newPaginatedNotesResponse(*paginatedNotes))
}

func (nh noteHandler) fetchNotesByNotebooks(w http.ResponseWriter, r *http.Request) {
	notebookIDs, ok := readIntQueryParams(r, pathParamNotebookID)
	if !ok {
		writeBadRequest(w, r.Header, newErrorResponse("invalid notebook ID parameters"))
		return
	}

	notebookNotesMap, err := nh.noteService.FetchNotesByNotebooks(notebookIDs)
	if err != nil {
		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeOK(w, r.Header, newNotebookNotesResponse(notebookNotesMap))
}

func (nh noteHandler) fetchNotesBySearchQuery(w http.ResponseWriter, r *http.Request) {
	searchQuery := readStringQueryParam(r, "query")
	page := readPaginationParam(r)

	paginatedNotes, err := nh.noteService.FetchNotesBySearchQuery(searchQuery, page)
	if err != nil {
		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeOK(w, r.Header, newPaginatedNotesResponse(*paginatedNotes))
}

func (nh noteHandler) updateNote(w http.ResponseWriter, r *http.Request) {
	noteID, _ := readIntPathParam(r, pathParamNoteID)

	var reqBody updateNoteRequestBody
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeBadRequest(w, r.Header, newErrorResponse(err.Error()))
		return
	}

	note, err := nh.noteService.UpdateNote(service.UpdateNoteParameters{
		NoteID:     noteID,
		Name:       reqBody.Name,
		Content:    reqBody.Content,
		NotebookID: reqBody.NotebookID,
	})
	if err != nil {
		var dErr *storage.DuplicateError
		if errors.As(err, &dErr) {
			writeBadRequest(w, r.Header, newErrorResponse(dErr.Error()))
			return
		}

		var rErr *storage.ReferenceError
		if errors.As(err, &rErr) {
			writeBadRequest(w, r.Header, newErrorResponse(rErr.Error()))
			return
		}

		var nfErr *storage.NotFoundError
		if errors.As(err, &nfErr) {
			writeBadRequest(w, r.Header, newErrorResponse(nfErr.Error()))
			return
		}

		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeCreated(w, r.Header, newNoteResponse(*note))
}

func (nh noteHandler) deleteNote(w http.ResponseWriter, r *http.Request) {
	noteID, _ := readIntPathParam(r, pathParamNoteID)

	if err := nh.noteService.DeleteNote(noteID); err != nil {
		var nfErr *storage.NotFoundError
		if errors.As(err, &nfErr) {
			writeNotFound(w, r.Header, newErrorResponse(nfErr.Error()))
			return
		}

		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeNoContent(w)
}
