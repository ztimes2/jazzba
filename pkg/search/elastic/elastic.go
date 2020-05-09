package elastic

import (
	"context"

	"github.com/olivere/elastic/v7"
)

type notebook struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type note struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	Notebook notebook `json:"notebook"`
}

const (
	indexNotes = "notes"

	fieldNoteName     = "name"
	fieldNoteContent  = "content"
	fieldNoteTags     = "tags"
	fieldNotebook     = "notebook"
	fieldNotebookID   = "notebook.id"
	fieldNotebookName = "notebook.name"
)

// Config holds configuration for connecting to ElasticSearch.
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

func (c Config) toURL() string {
	return "http://" + c.Host + ":" + c.Port
}

// NewClient initializes a new ElasticSearch client.
func NewClient(cfg Config) (*elastic.Client, error) {
	return elastic.NewClient(
		// TODO: enable basic auth
		elastic.SetURL(cfg.toURL()),
		elastic.SetHealthcheck(true),
		elastic.SetSniff(false),
	)
}

// InitIndices initializes all neccessary indices in ElasticSearch if they do
// not exist.
func InitIndices(client *elastic.Client) error {
	exists, err := client.IndexExists(indexNotes).Do(context.Background())
	if err != nil {
		return err
	}
	if !exists {
		_, err := client.CreateIndex(indexNotes).Do(context.Background())
		return err
	}
	return nil
}
