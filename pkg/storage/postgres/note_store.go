package postgres

import (
	"database/sql"

	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// TODO: make NoteStore drier

// NoteStore implements the storage.NoteStore interface and provides functionality
// for performing CRUD operations with notes in a PostgreSQL database.
type NoteStore struct {
	db *sql.DB
}

// NewNoteStore initializes a new instance of NoteStore.
func NewNoteStore(db *sql.DB) *NoteStore {
	return &NoteStore{
		db: db,
	}
}

// BeginTx initializes a new PostgreSQL database transaction and returns it.
func (ns NoteStore) BeginTx() (storage.Tx, error) {
	return ns.db.Begin()
}

var sqlQueryBaseSelectFromNotes = sqlQueryBase.
	Select("notes.id", "notes.name", "notes.content", "notes.notebook_id",
		"notes.created_at", "notes.updated_at").
	From("notes")

// CreateOne creates a new note in a PostgreSQL database using note parameters
// within the scope of a transaction.
func (ns NoteStore) CreateOne(tx storage.Tx, params storage.CreateNoteParameters,
) (*storage.Note, error) {

	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errNonSQLTransaction
	}

	query, args, err := sqlQueryBase.
		Insert("notes").
		Columns("name", "content", "notebook_id").
		Values(params.Name, params.Content, params.NotebookID).
		Suffix("RETURNING id, name, content, notebook_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	var note storage.Note
	if err := sqlTx.QueryRow(query, args...).Scan(
		&note.ID,
		&note.Name,
		&note.Content,
		&note.NotebookID,
		&note.CreatedAt,
		&note.UpdatedAt,
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

	return &note, nil
}

// FetchOne fetches a note from a PostgreSQL database by its ID.
func (ns NoteStore) FetchOne(noteID int) (*storage.Note, error) {
	query, args, err := sqlQueryBaseSelectFromNotes.
		Where(squirrel.Eq{"notes.id": noteID}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	var note storage.Note
	if err := ns.db.QueryRow(query, args...).Scan(
		&note.ID,
		&note.Name,
		&note.Content,
		&note.NotebookID,
		&note.CreatedAt,
		&note.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.NewNotFoundError()
		}
		return nil, newSQLQueryExecutionError(err)
	}

	return &note, nil
}

// FetchAllPaginated fetches all notes from a PostgreSQL database and paginates
// them using a limit and an offset.
func (ns NoteStore) FetchAllPaginated(limit, offset int) ([]storage.Note, error) {
	query, args, err := sqlQueryBaseSelectFromNotes.
		Limit(uint64(limit)).
		Offset(uint64(limit)).
		OrderBy("notes.created_at DESC").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	rows, err := ns.db.Query(query, args...)
	if err != nil {
		return nil, newSQLQueryExecutionError(err)
	}
	defer rows.Close()

	var notes []storage.Note
	for rows.Next() {
		var note storage.Note
		if err := rows.Scan(
			&note.ID,
			&note.Name,
			&note.Content,
			&note.NotebookID,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, newSQLRowScanError(err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// FetchMany fetches notes from a PostgreSQL database by their IDs.
func (ns NoteStore) FetchMany(noteIDs []int) ([]storage.Note, error) {
	query, args, err := sqlQueryBaseSelectFromNotes.
		Where(squirrel.Eq{"notes.id": noteIDs}).
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	rows, err := ns.db.Query(query, args...)
	if err != nil {
		return nil, newSQLQueryExecutionError(err)
	}
	defer rows.Close()

	var notes []storage.Note
	for rows.Next() {
		var note storage.Note
		if err := rows.Scan(
			&note.ID,
			&note.Name,
			&note.Content,
			&note.NotebookID,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, newSQLRowScanError(err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// FetchManyByNotebookPaginated fetches notes from a PostgreSQL database by ID
// of a notebook they are associated with and paginates them using a limit and
// an offset.
func (ns NoteStore) FetchManyByNotebookPaginated(notebookID, limit, offset int,
) ([]storage.Note, error) {

	query, args, err := sqlQueryBaseSelectFromNotes.
		Where(squirrel.Eq{"notes.notebook_id": notebookID}).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		OrderBy("notes.created_at DESC").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	rows, err := ns.db.Query(query, args...)
	if err != nil {
		return nil, newSQLQueryExecutionError(err)
	}
	defer rows.Close()

	var notes []storage.Note
	for rows.Next() {
		var note storage.Note
		if err := rows.Scan(
			&note.ID,
			&note.Name,
			&note.Content,
			&note.NotebookID,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, newSQLRowScanError(err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// FetchManyByNotebooks fetches notes from a PostgreSQL database by IDs of
// notebooks they are associated with and returns them as a map.
func (ns NoteStore) FetchManyByNotebooks(notebookIDs []int,
) (storage.NotebookNotesMap, error) {

	query, args, err := sqlQueryBaseSelectFromNotes.
		Where(squirrel.Eq{"notes.notebook_id": notebookIDs}).
		OrderBy("notes.created_at DESC").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	rows, err := ns.db.Query(query, args...)
	if err != nil {
		return nil, newSQLQueryExecutionError(err)
	}
	defer rows.Close()

	notebookNotesMap := storage.NotebookNotesMap{}
	for rows.Next() {
		var note storage.Note
		if err := rows.Scan(
			&note.ID,
			&note.Name,
			&note.Content,
			&note.NotebookID,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, newSQLRowScanError(err)
		}
		notebookNotesMap[note.NotebookID] = append(
			notebookNotesMap[note.NotebookID], note)
	}

	return notebookNotesMap, nil
}

// UpdateOne updates an existing note in a PostgreSQL database using note
// parameters within the scope of a transaction.
func (ns NoteStore) UpdateOne(tx storage.Tx, params storage.UpdateNoteParameters,
) (*storage.Note, error) {

	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errNonSQLTransaction
	}

	query, args, err := sqlQueryBase.
		Update("notes").
		SetMap(map[string]interface{}{
			"name":        params.Name,
			"content":     params.Content,
			"notebook_id": params.NotebookID,
			// FIXME can be improved by adding a trigger on the database end.
			"updated_at": squirrel.Expr("NOW()"),
		}).
		Where(squirrel.Eq{"id": params.NoteID}).
		Suffix("RETURNING id, name, content, notebook_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	var note storage.Note
	if err := sqlTx.QueryRow(query, args...).Scan(
		&note.ID,
		&note.Name,
		&note.Content,
		&note.NotebookID,
		&note.CreatedAt,
		&note.UpdatedAt,
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.NewNotFoundError()
		}
		return nil, newSQLQueryExecutionError(err)
	}

	return &note, nil
}

// DeleteOne deletes an existing note from a PostgreSQL database by its ID within
// the scope of a transaction.
func (ns NoteStore) DeleteOne(tx storage.Tx, noteID int) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errNonSQLTransaction
	}

	query, args, err := sqlQueryBase.
		Delete("notes").
		Where(squirrel.Eq{"id": noteID}).
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
		return newSQLAffectedRowsReadError(err)
	}

	if rowsAffected == 0 {
		return storage.NewNotFoundError()
	}

	return nil
}
