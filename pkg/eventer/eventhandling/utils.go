package eventhandling

import (
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/pkg/errors"
)

func newPayloadDecodingError(err error) error {
	return errors.Wrap(err, "could not decode event payload")
}

func extractTagsFromNoteTags(noteTags []storage.NoteTag) []string {
	var extractedTags []string
	for _, noteTag := range noteTags {
		extractedTags = append(extractedTags, noteTag.TagName)
	}
	return extractedTags
}

func newNote(note storage.Note, notebook storage.Notebook,
	noteTags []storage.NoteTag) search.Note {
	return search.Note{
		ID:      note.ID,
		Name:    note.Name,
		Content: note.Content,
		Notebook: search.Notebook{
			ID:   notebook.ID,
			Name: notebook.Name,
		},
		Tags: extractTagsFromNoteTags(noteTags),
	}
}
