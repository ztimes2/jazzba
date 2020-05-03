package httphandling

import (
	"encoding/json"
	"errors"
	"jazzba/pkg/api/service"
	"jazzba/pkg/storage"
	"net/http"

	"github.com/sirupsen/logrus"
)

type notebookHandler struct {
	notebookService service.Notebooker
	logger          logrus.FieldLogger
}

func newNotebookHandler(notebookService service.Notebooker,
	logger logrus.FieldLogger) *notebookHandler {
	return &notebookHandler{
		notebookService: notebookService,
		logger:          logger,
	}
}

func (nh notebookHandler) createNotebook(w http.ResponseWriter, r *http.Request) {
	var reqBody createNotebookRequestBody
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeBadRequest(w, r.Header, newErrorResponse(err.Error()))
		return
	}

	notebook, err := nh.notebookService.CreateNotebook(
		service.CreateNotebookParameters{
			Name: reqBody.Name,
		},
	)
	if err != nil {
		var dpErr *storage.DuplicateResourceError
		if errors.As(err, &dpErr) {
			writeBadRequest(w, r.Header, newErrorResponse(dpErr.Error()))
			return
		}

		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeCreated(w, r.Header, newNotebookResponse(*notebook))
}

func (nh notebookHandler) fetchNotebook(w http.ResponseWriter, r *http.Request) {
	notebookID, _ := readIntPathParam(r, pathParamNotebookID)

	notebook, err := nh.notebookService.FetchNotebook(notebookID)
	if err != nil {
		var rnfErr *storage.ResourceNotFoundError
		if errors.As(err, &rnfErr) {
			writeNotFound(w, r.Header, newErrorResponse(rnfErr.Error()))
			return
		}

		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeOK(w, r.Header, newNotebookResponse(*notebook))
}

func (nh notebookHandler) fetchNotebooks(w http.ResponseWriter, r *http.Request) {
	page := readPaginationParam(r)

	paginatedNotebooks, err := nh.notebookService.FetchNotebooks(page)
	if err != nil {
		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeOK(w, r.Header, newPaginatedNotebooksResponse(*paginatedNotebooks))
}

func (nh notebookHandler) updateNotebook(w http.ResponseWriter, r *http.Request) {
	notebookID, _ := readIntPathParam(r, pathParamNotebookID)

	var reqBody updateNotebookRequestBody
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeBadRequest(w, r.Header, newErrorResponse(err.Error()))
		return
	}

	notebook, err := nh.notebookService.UpdateNotebook(
		service.UpdateNotebookParameters{
			NotebookID: notebookID,
			Name:       reqBody.Name,
		},
	)
	if err != nil {
		var drErr *storage.DuplicateResourceError
		if errors.As(err, &drErr) {
			writeBadRequest(w, r.Header, newErrorResponse(drErr.Error()))
			return
		}

		var rnfErr *storage.ResourceNotFoundError
		if errors.As(err, &rnfErr) {
			writeNotFound(w, r.Header, newErrorResponse(rnfErr.Error()))
			return
		}

		writeInternalServerError(w, nh.logger, err)
		return
	}

	writeCreated(w, r.Header, newNotebookResponse(*notebook))
}

func (nh notebookHandler) deleteNotebook(w http.ResponseWriter, r *http.Request) {
	notebookID, _ := readIntPathParam(r, pathParamNotebookID)

	if err := nh.notebookService.DeleteNotebook(notebookID); err != nil {
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
