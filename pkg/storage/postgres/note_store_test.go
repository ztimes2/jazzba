package postgres

import (
	"github.com/ztimes2/jazzba/pkg/storage"
)

var _ storage.NoteStore = (*NoteStore)(nil)

// TODO: add tests
