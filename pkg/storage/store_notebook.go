package storage

import "time"

// Notebook represents a notebook in a data store.
type Notebook struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpdateNotebookParameters holds parameters for updating a notebook in a data
// store.
type UpdateNotebookParameters struct {
	NotebookID int
	Name       string
}

// NotebookStore abstracts functionality for performing CRUD operations with
// notebooks in a data store.
type NotebookStore interface {
	Transactor
	CreateOne(tx Tx, notebookName string) (*Notebook, error)
	FetchOne(notebookID int) (*Notebook, error)
	FetchAllPaginated(limit, offset int) ([]Notebook, error)
	UpdateOne(tx Tx, p UpdateNotebookParameters) (*Notebook, error)
	DeleteOne(tx Tx, notebookID int) error
}
