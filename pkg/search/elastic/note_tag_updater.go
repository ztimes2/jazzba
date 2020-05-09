package elastic

import (
	"context"
	"strconv"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// NoteTagUpdater implements search.NoteTagUpdater interface and provides
// functionality for updating note tags in ElasticSearch.
type NoteTagUpdater struct {
	client *elastic.Client
}

// NewNoteTagUpdater initializes a new instance of NoteTagUpdater.
func NewNoteTagUpdater(client *elastic.Client) *NoteTagUpdater {
	return &NoteTagUpdater{
		client: client,
	}
}

// UpdateMany updates note tags in ElasticSearch.
func (ntu NoteTagUpdater) UpdateMany(noteID int, tags []string) error {
	if _, err := ntu.client.Update().
		Index(indexNotes).
		Id(strconv.Itoa(noteID)).
		Doc(map[string]interface{}{
			fieldNoteTags: tags,
		}).
		Do(context.Background()); err != nil {
		return errors.Wrap(err, "could not update document in index")
	}
	return nil
}
