package elastic

import (
	"context"
	"encoding/json"

	"github.com/ztimes2/jazzba/pkg/search"

	"github.com/sirupsen/logrus"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// NoteSearcher implements search.NoteSearcher interface and provides functionality
// for searching notes using ElasticSearch.
type NoteSearcher struct {
	client *elastic.Client
	logger logrus.FieldLogger
}

// NewNoteSearcher initializes a new instance of NoteSearcher.
func NewNoteSearcher(client *elastic.Client) *NoteSearcher {
	return &NoteSearcher{
		client: client,
	}
}

func convertSearchHitToSearchNote(hit *elastic.SearchHit) (*search.Note, error) {
	var n note
	if err := json.Unmarshal(hit.Source, &n); err != nil {
		return nil, errors.Wrap(err, "could not decode document "+hit.Id)
	}
	return &search.Note{
		ID:      n.ID,
		Name:    n.Name,
		Content: n.Content,
		Notebook: search.Notebook{
			ID:   n.Notebook.ID,
			Name: n.Notebook.Name,
		},
		Tags: n.Tags,
	}, nil
}

const (
	searchTypeBestFields = "best_fields"
	operatorOr           = "or"
)

// SearchByQuery performs a full-text search for notes using ElasticSearch.
func (ns NoteSearcher) SearchByQuery(query string, limit, offset int,
) ([]search.Note, error) {

	multiMatchQuery := elastic.NewMultiMatchQuery(
		query,
		fieldNoteName,
		fieldNoteContent,
		fieldNoteTags,
		fieldNotebookName).
		Type(searchTypeBestFields).
		Operator(operatorOr)

	res, err := ns.client.
		Search(indexNotes).
		Query(multiMatchQuery).
		Size(limit).
		From(offset).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "could not execute search request")
	}

	var notes []search.Note
	for _, hit := range res.Hits.Hits {
		n, err := convertSearchHitToSearchNote(hit)
		if err != nil {
			return nil, err
		}
		notes = append(notes, *n)
	}
	return notes, nil
}
