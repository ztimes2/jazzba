package storage

import "time"

// Note represents a note in a data store.
type Note struct {
	ID         int
	Name       string
	Content    string
	NotebookID int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CreateNoteParameters holds parameters for creating a note in a data store.
type CreateNoteParameters struct {
	Transaction Tx
	Name        string
	Content     string
	NotebookID  int
}

// UpdateNoteParameters holds parameters for updating a note in a data store.
type UpdateNoteParameters struct {
	Transaction Tx
	NoteID      int
	Name        string
	Content     string
	NotebookID  int
}

// NotebookNotesMap represents a map where a key is a notebook ID and a value is
// a collection of notes associated with that particular notebook.
type NotebookNotesMap map[int][]Note

// NoteStore abstracts functionality for performing CRUD operations with notes
// in a data store.
type NoteStore interface {
	Transactor
	CreateOne(CreateNoteParameters) (*Note, error)
	FetchOne(noteID int) (*Note, error)
	FetchAllPaginated(limit, offset int) ([]Note, error)
	FetchMany(noteIDs []int) ([]Note, error)
	FetchManyByNotebookPaginated(notebookID, limit, offset int) ([]Note, error)
	FetchManyByNotebooks(notebookIDs []int) (NotebookNotesMap, error)
	UpdateOne(UpdateNoteParameters) (*Note, error)
	DeleteOne(tx Tx, noteID int) error
}
