package postgres

import (
	"database/sql"

	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// TODO: make NotebookStore drier

// NotebookStore implements the storage.NotebookStore interface and provides
// functionality for performing CRUD operations with notebooks in a PostgreSQL
// database.
type NotebookStore struct {
	db *sql.DB
}

// NewNotebookStore initializes a new instance of NotebookStore.
func NewNotebookStore(db *sql.DB) *NotebookStore {
	return &NotebookStore{
		db: db,
	}
}

// BeginTx initializes a new PostgreSQL database transaction and returns it.
func (ns NotebookStore) BeginTx() (storage.Tx, error) {
	return ns.db.Begin()
}

var sqlQueryBaseSelectFromNotebooks = sqlQueryBase.
	Select("notebooks.id", "notebooks.name", "notebooks.created_at",
		"notebooks.updated_at").
	From("notebooks")

// CreateOne creates a new notebook in a PostgreSQL database using notebook
// parameters within the scope of a transaction.
func (ns NotebookStore) CreateOne(tx storage.Tx, notebookName string,
) (*storage.Notebook, error) {

	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errNonSQLTransaction
	}

	query, args, err := sqlQueryBase.
		Insert("notebooks").
		Columns("name").
		Values(notebookName).
		Suffix("RETURNING id, name, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	var notebook storage.Notebook
	if err := sqlTx.QueryRow(query, args...).Scan(
		&notebook.ID,
		&notebook.Name,
		&notebook.CreatedAt,
		&notebook.UpdatedAt,
	); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errorCodeUniqueViolation {
			return nil, storage.NewDuplicateError()
		}
		return nil, newSQLQueryExecutionError(err)
	}

	return &notebook, nil
}

// FetchOne fetches a notebook from a PostgreSQL database by its ID.
func (ns NotebookStore) FetchOne(notebookID int) (*storage.Notebook, error) {
	query, args, err := sqlQueryBaseSelectFromNotebooks.
		Where(squirrel.Eq{"notebooks.id": notebookID}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	var notebook storage.Notebook
	if err := ns.db.QueryRow(query, args...).Scan(
		&notebook.ID,
		&notebook.Name,
		&notebook.CreatedAt,
		&notebook.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.NewNotFoundError()
		}
		return nil, newSQLQueryExecutionError(err)
	}

	return &notebook, nil
}

// FetchAllPaginated fetches all notebooks from a PostgreSQL database and
// paginates them using a limit and an offset.
func (ns NotebookStore) FetchAllPaginated(limit, offset int,
) ([]storage.Notebook, error) {

	query, args, err := sqlQueryBaseSelectFromNotebooks.
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		OrderBy("notebooks.created_at DESC").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	rows, err := ns.db.Query(query, args...)
	if err != nil {
		return nil, newSQLQueryExecutionError(err)
	}
	defer rows.Close()

	var notebooks []storage.Notebook
	for rows.Next() {
		var notebook storage.Notebook
		if err := rows.Scan(
			&notebook.ID,
			&notebook.Name,
			&notebook.CreatedAt,
			&notebook.UpdatedAt,
		); err != nil {
			return nil, newSQLRowScanError(err)
		}
		notebooks = append(notebooks, notebook)
	}

	return notebooks, nil
}

// UpdateOne updates an existing notebook in a PostgreSQL database using notebook
// parameters within the scope of a transaction.
func (ns NotebookStore) UpdateOne(tx storage.Tx,
	params storage.UpdateNotebookParameters) (*storage.Notebook, error) {

	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errNonSQLTransaction
	}

	query, args, err := sqlQueryBase.
		Update("notebooks").
		SetMap(map[string]interface{}{
			"name": params.Name,
			// FIXME can be improved by adding a trigger on the database end.
			"updated_at": squirrel.Expr("NOW()"),
		}).
		Where(squirrel.Eq{"id": params.NotebookID}).
		Suffix("RETURNING id, name, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, newSQLQueryBuilderError(err)
	}

	var notebook storage.Notebook
	if err := sqlTx.QueryRow(query, args...).Scan(
		&notebook.ID,
		&notebook.Name,
		&notebook.CreatedAt,
		&notebook.UpdatedAt,
	); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errorCodeUniqueViolation {
			return nil, storage.NewDuplicateError()
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.NewNotFoundError()
		}
		return nil, newSQLQueryExecutionError(err)
	}

	return &notebook, nil
}

// DeleteOne deletes an existing notebook from a PostgreSQL database by its ID
// within the scope of a transaction.
func (ns NotebookStore) DeleteOne(tx storage.Tx, notebookID int) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errNonSQLTransaction
	}

	query, args, err := sqlQueryBase.
		Delete("notebooks").
		Where(squirrel.Eq{"id": notebookID}).
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
