package postgres

import (
	"github.com/ztimes2/jazzba/pkg/storage"
)

var _ storage.NoteTagStore = (*NoteTagStore)(nil)

// TODO: add tests
