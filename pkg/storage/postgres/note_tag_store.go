package postgres

import (
	"database/sql"

	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// TODO: make NoteTagStore drier

// NoteTagStore implements the storage.NoteTagStore interface and provides
// provides functionality for performing CRUD operations with note tags in a
// PostgreSQL database.
type NoteTagStore struct {
	db *sql.DB
}

// NewNoteTagStore initializes a new instance of NoteTagStore.
func NewNoteTagStore(db *sql.DB) *NoteTagStore {
	return &NoteTagStore{
		db: db,
	}
}

var sqlQueryBaseSelectFromNoteTags = sqlQueryBase.
	Select("note_tags.tag_name", "note_tags.note_id", "note_tags.created_at",
		"note_tags.updated_at").
	From("note_tags")

// BeginTx initializes a new PostgreSQL database transaction and returns it.
func (nts NoteTagStore) BeginTx() (storage.Tx, error) {
	return nts.db.Begin()
}

// CreateOne creates a new note tag in a PostgreSQL database within the scope of
// a transaction.
func (nts NoteTagStore) CreateOne(tx storage.Tx,
	params storage.CreateNoteTagParameters) (*storage.NoteTag, error) {

	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errNonSQLTransaction
	}

	query, args, err := sqlQueryBase.
		Insert("note_tags").
		Columns("tag_name", "note_id").
		Values(params.TagName, params.NoteID).
		Suffix("RETURNING tag_name, note_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	var noteTag storage.NoteTag
	if err := sqlTx.QueryRow(query, args...).Scan(
		&noteTag.TagName,
		&noteTag.NoteID,
		&noteTag.CreatedAt,
		&noteTag.UpdatedAt,
	); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case errorCodeUniqueViolation:
				return nil, storage.NewDuplicateError()
			case errorCodeForeignKeyViolation:
				return nil, storage.NewReferenceError()
			}
		}
		return nil, newSQLQueryExecutionError(err)
	}

	return &noteTag, nil
}

// FetchManyByNotePaginated fetches note tags from a PostgreSQL database by an
// ID of a note they are associated with and paginates them using a limit and an
// offset.
func (nts NoteTagStore) FetchManyByNotePaginated(noteID, limit, offset int,
) ([]storage.NoteTag, error) {

	query, args, err := sqlQueryBaseSelectFromNoteTags.
		Where(squirrel.Eq{"note_tags.note_id": noteID}).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		OrderBy("note_tags.created_at DESC").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	rows, err := nts.db.Query(query, args...)
	if err != nil {
		return nil, newSQLQueryExecutionError(err)
	}
	defer rows.Close()

	var noteTags []storage.NoteTag
	for rows.Next() {
		var noteTag storage.NoteTag
		if err := rows.Scan(
			&noteTag.TagName,
			&noteTag.NoteID,
			&noteTag.CreatedAt,
			&noteTag.UpdatedAt,
		); err != nil {
			return nil, newSQLRowScanError(err)
		}
		noteTags = append(noteTags, noteTag)
	}

	return noteTags, nil
}

// FetchManyByNote fetches note tags from a PostgreSQL database by an ID of a
// note they are associated with.
func (nts NoteTagStore) FetchManyByNote(noteID int) ([]storage.NoteTag, error) {
	query, args, err := sqlQueryBaseSelectFromNoteTags.
		Where(squirrel.Eq{"note_tags.note_id": noteID}).
		OrderBy("note_tags.created_at DESC").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	rows, err := nts.db.Query(query, args...)
	if err != nil {
		return nil, newSQLQueryExecutionError(err)
	}
	defer rows.Close()

	var noteTags []storage.NoteTag
	for rows.Next() {
		var noteTag storage.NoteTag
		if err := rows.Scan(
			&noteTag.TagName,
			&noteTag.NoteID,
			&noteTag.CreatedAt,
			&noteTag.UpdatedAt,
		); err != nil {
			return nil, newSQLRowScanError(err)
		}
		noteTags = append(noteTags, noteTag)
	}

	return noteTags, nil
}

// FetchManyByNotes fetches note tags from a PostgreSQL database by IDs of notes
// they are associated with and returns them as a map.
func (nts NoteTagStore) FetchManyByNotes(noteIDs []int,
) (storage.NoteTagsMap, error) {

	query, args, err := sqlQueryBaseSelectFromNoteTags.
		Where(squirrel.Eq{"note_tags.note_id": noteIDs}).
		OrderBy("note_tags.created_at DESC").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	rows, err := nts.db.Query(query, args...)
	if err != nil {
		return nil, newSQLQueryExecutionError(err)
	}
	defer rows.Close()

	noteTagsMap := storage.NoteTagsMap{}
	for rows.Next() {
		var noteTag storage.NoteTag
		if err := rows.Scan(
			&noteTag.TagName,
			&noteTag.NoteID,
			&noteTag.CreatedAt,
			&noteTag.UpdatedAt,
		); err != nil {
			return nil, newSQLRowScanError(err)
		}
		noteTagsMap[noteTag.NoteID] = append(noteTagsMap[noteTag.NoteID], noteTag)
	}

	return noteTagsMap, nil
}

// DeleteOne deletes an existing note tag from a PostgreSQL database within the
// scope of a transaction.
func (nts NoteTagStore) DeleteOne(tx storage.Tx,
	params storage.DeleteNoteTagParameters) error {

	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errNonSQLTransaction
	}

	query, args, err := sqlQueryBase.
		Delete("note_tags").
		Where(squirrel.Eq{"tag_name": params.TagName, "note_id": params.NoteID}).
		ToSql()
	if err != nil {
		return newSQLQueryBuilderError(err)
	}

	res, err := sqlTx.Exec(query, args...)
	if err != nil {
		return newSQLQueryExecutionError(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return newSQLRowScanError(err)
	}

	if rowsAffected == 0 {
		return storage.NewNotFoundError()
	}

	return nil
}
