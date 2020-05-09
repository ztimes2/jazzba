package elastic

import (
	"context"
	"strconv"

	"github.com/ztimes2/jazzba/pkg/search"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// NoteUpdater implements search.NoteUpdater interface and provides functionality
// for updating notes in ElasticSearch.
type NoteUpdater struct {
	client *elastic.Client
}

// NewNoteUpdater initializes a new instance of NoteUpdater.
func NewNoteUpdater(client *elastic.Client) *NoteUpdater {
	return &NoteUpdater{
		client: client,
	}
}

// CreateOne creates a note in ElasticSearch.
func (nu NoteUpdater) CreateOne(n search.Note) error {
	if _, err := nu.client.Index().
		Index(indexNotes).
		Id(strconv.Itoa(n.ID)).
		BodyJson(note{
			ID:      n.ID,
			Name:    n.Name,
			Content: n.Content,
			Notebook: notebook{
				ID:   n.Notebook.ID,
				Name: n.Notebook.Name,
			},
			Tags: n.Tags,
		}).
		Do(context.Background()); err != nil {
		return errors.Wrap(err, "could not create document in index")
	}
	return nil
}

// UpdateOne updates a note in ElasticSearch.
func (nu NoteUpdater) UpdateOne(n search.Note) error {
	if _, err := nu.client.Update().
		Index(indexNotes).
		Id(strconv.Itoa(n.ID)).
		Doc(note{
			ID:      n.ID,
			Name:    n.Name,
			Content: n.Content,
			Notebook: notebook{
				ID:   n.Notebook.ID,
				Name: n.Notebook.Name,
			},
			Tags: n.Tags,
		}).
		Do(context.Background()); err != nil {
		return errors.Wrap(err, "could not update document in index")
	}
	return nil
}

// DeleteOne deletes a note from ElasticSearch.
func (nu NoteUpdater) DeleteOne(noteID int) error {
	if _, err := nu.client.
		Delete().
		Index(indexNotes).
		Id(strconv.Itoa(noteID)).
		Do(context.Background()); err != nil {
		return errors.Wrap(err, "could not delete document from index")
	}
	return nil
}
