package httphandling

import (
	"encoding/json"
	"time"

	"github.com/ztimes2/jazzba/pkg/api/p8n"
	"github.com/ztimes2/jazzba/pkg/api/service"
	"github.com/ztimes2/jazzba/pkg/storage"
)

// response is an abstract API response used for writing an HTTP response body.
type response interface {
	toJSON() ([]byte, error)
}

func timeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

// errorResponse represents a general error response.
type errorResponse struct {
	Message string `json:"message"`
}

func newErrorResponse(message string) errorResponse {
	return errorResponse{
		Message: message,
	}
}

func (er errorResponse) toJSON() ([]byte, error) {
	return json.Marshal(&er)
}

// page represents a page defined by its limit and offset. It is meant to be used
// for constucting paginated responses.
type page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// pagination holds next and previous pages for pagination. It is meant to be
// used by paginated responses.
type pagination struct {
	Next     *page `json:"next"`
	Previous *page `json:"previous"`
}

func newPagination(p p8n.Pagination) pagination {
	var next, previous *page

	if p.HasNextPage() {
		next = &page{
			Limit:  p.Next.Limit,
			Offset: p.Next.Offset,
		}
	}

	if p.HasPreviousPage() {
		previous = &page{
			Limit:  p.Previous.Limit,
			Offset: p.Previous.Offset,
		}
	}

	return pagination{
		Next:     next,
		Previous: previous,
	}
}

// notebookResponse represents a notebook response.
type notebookResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func newNotebookResponse(notebook storage.Notebook) notebookResponse {
	return notebookResponse{
		ID:        notebook.ID,
		Name:      notebook.Name,
		CreatedAt: timeToString(notebook.CreatedAt),
		UpdatedAt: timeToString(notebook.UpdatedAt),
	}
}

func newNotebooksResponses(notebooks []storage.Notebook) []notebookResponse {
	var notebookResponses []notebookResponse
	for _, notebook := range notebooks {
		notebookResponses = append(notebookResponses, newNotebookResponse(notebook))
	}
	return notebookResponses
}

func (nr notebookResponse) toJSON() ([]byte, error) {
	return json.Marshal(&nr)
}

// paginatedNotebooksResponse represents a response containing a paginated list
// of notebooks.
type paginatedNotebooksResponse struct {
	Notebooks  []notebookResponse `json:"notebooks"`
	Pagination pagination         `json:"pagination"`
}

func newPaginatedNotebooksResponse(paginatedNotebooks service.PaginatedNotebooks,
) paginatedNotebooksResponse {
	return paginatedNotebooksResponse{
		Notebooks:  newNotebooksResponses(paginatedNotebooks.Notebooks),
		Pagination: newPagination(paginatedNotebooks.Pagination),
	}
}

func (pnr paginatedNotebooksResponse) toJSON() ([]byte, error) {
	return json.Marshal(&pnr)
}

// noteResponse represents a note response.
type noteResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	NotebookID int    `json:"notebook_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func newNoteResponse(note storage.Note) noteResponse {
	return noteResponse{
		ID:         note.ID,
		Name:       note.Name,
		Content:    note.Content,
		NotebookID: note.NotebookID,
		CreatedAt:  timeToString(note.CreatedAt),
		UpdatedAt:  timeToString(note.UpdatedAt),
	}
}

func newNoteResponses(notes []storage.Note) []noteResponse {
	var noteResponses []noteResponse
	for _, note := range notes {
		noteResponses = append(noteResponses, newNoteResponse(note))
	}
	return noteResponses
}

func (nr noteResponse) toJSON() ([]byte, error) {
	return json.Marshal(&nr)
}

// paginatedNotesResponse represents a response containing a paginated list of
// notes.
type paginatedNotesResponse struct {
	Notes      []noteResponse `json:"notes"`
	Pagination pagination     `json:"pagination"`
}

func newPaginatedNotesResponse(paginatedNotes service.PaginatedNotes,
) paginatedNotesResponse {
	return paginatedNotesResponse{
		Notes:      newNoteResponses(paginatedNotes.Notes),
		Pagination: newPagination(paginatedNotes.Pagination),
	}
}

func (pnr paginatedNotesResponse) toJSON() ([]byte, error) {
	return json.Marshal(&pnr)
}

// notebookNotesResponse represents a response containing a mapping of notebooks
// and notes associated with them.
type notebookNotesResponse struct {
	NotebookNotesMap map[int][]noteResponse `json:"notebook_notes_map"`
}

func newNotebookNotesResponse(notebookNotesMap storage.NotebookNotesMap,
) notebookNotesResponse {

	if len(notebookNotesMap) == 0 {
		return notebookNotesResponse{}
	}

	notebookNotesResponse := notebookNotesResponse{
		NotebookNotesMap: make(map[int][]noteResponse),
	}
	for notebookID, notes := range notebookNotesMap {
		notebookNotesResponse.NotebookNotesMap[notebookID] = newNoteResponses(notes)
	}
	return notebookNotesResponse
}

func (nnr notebookNotesResponse) toJSON() ([]byte, error) {
	return json.Marshal(&nnr)
}

// noteTagResponse represents a note tag response.
type noteTagResponse struct {
	NoteID    int    `json:"note_id"`
	TagName   string `json:"tag_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func newNoteTagResponse(noteTag storage.NoteTag) noteTagResponse {
	return noteTagResponse{
		NoteID:    noteTag.NoteID,
		TagName:   noteTag.TagName,
		CreatedAt: timeToString(noteTag.CreatedAt),
		UpdatedAt: timeToString(noteTag.UpdatedAt),
	}
}

func newNoteTagResponses(noteTags []storage.NoteTag) []noteTagResponse {
	var noteTagResponses []noteTagResponse
	for _, noteTag := range noteTags {
		noteTagResponses = append(noteTagResponses, newNoteTagResponse(noteTag))
	}
	return noteTagResponses
}

func (ntr noteTagResponse) toJSON() ([]byte, error) {
	return json.Marshal(&ntr)
}

// paginatedNoteTagsResponse represents a paginated note tags response.
type paginatedNoteTagsResponse struct {
	NoteTags   []noteTagResponse `json:"note_tags"`
	Pagination pagination        `json:"pagination"`
}

func newPaginatedNoteTagsResponse(paginatedNoteTags service.PaginatedNoteTags,
) paginatedNoteTagsResponse {
	return paginatedNoteTagsResponse{
		NoteTags:   newNoteTagResponses(paginatedNoteTags.NoteTags),
		Pagination: newPagination(paginatedNoteTags.Pagination),
	}
}

func (pntr paginatedNoteTagsResponse) toJSON() ([]byte, error) {
	return json.Marshal(&pntr)
}

// noteTagsResponse represents a note tags response containing a mapping of
// notebooks and notes associated with them.
type noteTagsResponse struct {
	NoteTagsMap map[int][]noteTagResponse `json:"note_tags_map"`
}

func newNoteTagsResponse(noteTagsMap storage.NoteTagsMap) noteTagsResponse {
	if len(noteTagsMap) == 0 {
		return noteTagsResponse{}
	}

	noteTagsResponse := noteTagsResponse{
		NoteTagsMap: make(map[int][]noteTagResponse),
	}
	for noteID, tags := range noteTagsMap {
		noteTagsResponse.NoteTagsMap[noteID] = newNoteTagResponses(tags)
	}
	return noteTagsResponse
}

func (ntr noteTagsResponse) toJSON() ([]byte, error) {
	return json.Marshal(&ntr)
}
