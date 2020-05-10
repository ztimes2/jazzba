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
		var drErr *storage.DuplicateResourceError
		if errors.As(err, &drErr) {
			writeBadRequest(w, r.Header, newErrorResponse(drErr.Error()))
			return
		}

		var isrrErr *storage.InvalidSubResourceReferenceError
		if errors.As(err, &isrrErr) {
			writeBadRequest(w, r.Header, newErrorResponse(isrrErr.Error()))
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
		var rnfError *storage.ResourceNotFoundError
		if errors.As(err, &rnfError) {
			writeBadRequest(w, r.Header, newErrorResponse(rnfError.Error()))
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
	notebookIDs, ok := readIntQueryParams(r, "notebook_id")
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
		var drErr *storage.DuplicateResourceError
		if errors.As(err, &drErr) {
			writeBadRequest(w, r.Header, newErrorResponse(drErr.Error()))
			return
		}

		var isrrErr *storage.InvalidSubResourceReferenceError
		if errors.As(err, &isrrErr) {
			writeBadRequest(w, r.Header, newErrorResponse(isrrErr.Error()))
			return
		}

		var rnfErr *storage.ResourceNotFoundError
		if errors.As(err, &rnfErr) {
			writeBadRequest(w, r.Header, newErrorResponse(rnfErr.Error()))
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
		var rnfErr *storage.ResourceNotFoundError
		if errors.As(err, &rnfErr) {
			writeNotFound(w, r.Header, newErrorResponse(rnfErr.Error()))
			return
		}

		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeNoContent(w)
}
