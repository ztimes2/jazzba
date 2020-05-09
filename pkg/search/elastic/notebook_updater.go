package elastic

import (
	"context"
	"strconv"

	"github.com/ztimes2/jazzba/pkg/search"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// NotebookUpdater implements search.NotebookUpdater interface and provides
// functionality for updating notebooks in ElasticSearch.
type NotebookUpdater struct {
	client *elastic.Client
}

// NewNotebookUpdater initializes a new instance of NotebookUpdater.
func NewNotebookUpdater(client *elastic.Client) *NotebookUpdater {
	return &NotebookUpdater{
		client: client,
	}
}

// UpdateOne updates a notebook in ElasticSearch.
func (nu NotebookUpdater) UpdateOne(n search.Notebook) error {
	if _, err := nu.client.UpdateByQuery(indexNotes).
		Query(elastic.NewTermQuery(fieldNotebookID, n.ID)).
		Script(elastic.NewScriptInline(
			"ctx._source."+fieldNotebook+" = params.notebook").
			Param("notebook", notebook{
				ID:   n.ID,
				Name: n.Name,
			}),
		).
		Do(context.Background()); err != nil {
		return errors.Wrap(
			err, "could not update notes associated with notebook in index")
	}
	return nil
}

// DeleteOne deletes a notebook from ElasticSearch.
func (nu NotebookUpdater) DeleteOne(notebookID int) error {
	if _, err := nu.client.
		DeleteByQuery().
		Index(indexNotes).
		Query(
			elastic.NewTermQuery(fieldNotebookID, strconv.Itoa(notebookID)),
		).
		Do(context.Background()); err != nil {
		return errors.Wrap(
			err, "could not delete notes associated with notebook from index")
	}
	return nil
}
