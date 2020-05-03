package httphandling

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	headerKeyAccept      = "Accept"
	headerKeyContentType = "Content-Type"

	contentTypeJSON = "application/json"
)

type responseEncoderFunc func(response) ([]byte, error)

func toJSONResponse(r response) ([]byte, error) {
	return r.toJSON()
}

var responseEncoders = map[string]responseEncoderFunc{
	contentTypeJSON: toJSONResponse,
}

// writeResponse writes an HTTP status and a response body based on request
// headers to an HTTP response.
func writeResponse(w http.ResponseWriter, h http.Header, statusCode int, r response) {
	if r == nil {
		w.WriteHeader(statusCode)
		return
	}

	acceptHeader := h.Get(headerKeyAccept)

	encoderFunc, ok := responseEncoders[acceptHeader]
	if !ok {
		acceptHeader = contentTypeJSON
		encoderFunc = toJSONResponse
	}

	body, _ := encoderFunc(r)

	w.Header().Set(headerKeyContentType, acceptHeader)
	w.WriteHeader(statusCode)
	w.Write(body)
}

// writeOK writes the 200 OK HTTP status and a response body based on request
// headers to an HTTP response.
func writeOK(w http.ResponseWriter, h http.Header, r response) {
	writeResponse(w, h, http.StatusOK, r)
}

// writeCreated writes the 201 Created HTTP status and a response body based on
// request headers to an HTTP response.
func writeCreated(w http.ResponseWriter, h http.Header, r response) {
	writeResponse(w, h, http.StatusCreated, r)
}

// writeNoContent writes the 204 No Content HTTP status to the response.
func writeNoContent(w http.ResponseWriter) {
	writeResponse(w, nil, http.StatusNoContent, nil)
}

// writeInternalServerError writes the 500 Internal Server Error HTTP status to
// an HTTP response and logs an error using a logger.
func writeInternalServerError(w http.ResponseWriter, logger logrus.FieldLogger,
	err error) {
	w.WriteHeader(http.StatusInternalServerError)
}

// writeBadRequest writes the 403 Bad Request HTTP status and an error response
// body based on request headers to an HTTP response.
func writeBadRequest(w http.ResponseWriter, h http.Header, er errorResponse) {
	writeResponse(w, h, http.StatusBadRequest, er)
}

// writeNotFound writes the 404 Not Found HTTP status and an error response body
// based on request headers to an HTTP response.
func writeNotFound(w http.ResponseWriter, h http.Header, er errorResponse) {
	writeResponse(w, h, http.StatusNotFound, er)
}
