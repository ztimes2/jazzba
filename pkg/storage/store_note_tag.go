package storage

import "time"

// NoteTag represents a note tag in a data store.
type NoteTag struct {
	NoteID    int
	TagName   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateNoteTagParameters holds parameters for creating a note tag in a data
// store.
type CreateNoteTagParameters struct {
	NoteID  int
	TagName string
}

// DeleteNoteTagParameters holds parameters for deleting a note tag from a data
// store.
type DeleteNoteTagParameters struct {
	NoteID  int
	TagName string
}

// NoteTagsMap represents a map where a key is a note ID and a value is a
// collection of tags associated with that particular note.
type NoteTagsMap map[int][]NoteTag

// NoteTagStore abstracts functionality for performing CRUD operations with note
// tags in a data store.
type NoteTagStore interface {
	Transactor
	CreateOne(tx Tx, p CreateNoteTagParameters) (*NoteTag, error)
	FetchManyByNotePaginated(noteID, limit, offset int) ([]NoteTag, error)
	FetchManyByNote(noteID int) ([]NoteTag, error)
	FetchManyByNotes(noteIDs []int) (NoteTagsMap, error)
	DeleteOne(tx Tx, p DeleteNoteTagParameters) error
}
