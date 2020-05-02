package storage

// Tx abstracts a data store transaction.
type Tx interface {
	Commit() error
	Rollback() error
}

// Transactor abstracts initialization of a data store transaction. It is intended
// to be implemented and used by data stores which support transactions.
type Transactor interface {
	BeginTx() (Tx, error)
}
