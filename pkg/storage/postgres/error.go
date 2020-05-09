package postgres

import (
	"errors"
	"fmt"
)

const (
	errorCodeUniqueViolation     = "23505"
	errorCodeForeignKeyViolation = "23503"
)

var (
	errNonSQLTransaction = errors.New("transaction must be of type *sql.Tx")
)

type sqlQueryBuilderError struct {
	err error
}

func newSQLQueryBuilderError(err error) *sqlQueryBuilderError {
	return &sqlQueryBuilderError{
		err: err,
	}
}

func (e sqlQueryBuilderError) Error() string {
	return fmt.Sprintf("could not build sql query: %s", e.err)
}

type sqlQueryExecutionError struct {
	err error
}

func newSQLQueryExecutionError(err error) *sqlQueryExecutionError {
	return &sqlQueryExecutionError{
		err: err,
	}
}

func (e sqlQueryExecutionError) Error() string {
	return fmt.Sprintf("could not execute sql query: %s", e.err)
}

type sqlRowScanError struct {
	err error
}

func newSQLRowScanError(err error) *sqlRowScanError {
	return &sqlRowScanError{
		err: err,
	}
}

func (e sqlRowScanError) Error() string {
	return fmt.Sprintf("could not scan sql rows: %s", e.err)
}

type sqlAffectedRowsReadError struct {
	err error
}

func newSQLAffectedRowsReadError(err error) *sqlAffectedRowsReadError {
	return &sqlAffectedRowsReadError{
		err: err,
	}
}

func (e sqlAffectedRowsReadError) Error() string {
	return fmt.Sprintf("could not read affected sql rows: %s", e.err)
}
