package search

// Notebook represents a notebook in a search engine.
type Notebook struct {
	ID   int
	Name string
}

// Note represents a note in a search engine.
type Note struct {
	ID       int
	Name     string
	Content  string
	Notebook Notebook
	Tags     []string
}

// NoteSearcher abstracts functionality for searching notes using a search engine.
type NoteSearcher interface {
	SearchByQuery(query string, limit, offset int) ([]Note, error)
}

// NoteUpdater abstracts functionality for updating notes in a search engine.
type NoteUpdater interface {
	CreateOne(Note) error
	UpdateOne(Note) error
	DeleteOne(noteID int) error
}

// NoteTagUpdater abstracts functionality for updating note tags in a search
// engine.
type NoteTagUpdater interface {
	UpdateMany(noteID int, tags []string) error
}

// NotebookUpdater abstracts functionality for updating notebooks in a search
// engine.
type NotebookUpdater interface {
	UpdateOne(Notebook) error
	DeleteOne(notebookID int) error
}
