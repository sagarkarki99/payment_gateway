package db

import (
	"database/sql"
)

type TransactionRepository interface {
	Create(tx *Transaction) (*Transaction, error)
	Update(tx Transaction) error
	GetTransactionByGatewayTxnId(gatewayTxnId string) (*Transaction, error)
}

type SQLTransactionRepository struct {
	db *sql.DB
}

var NewTransactionRepository = func(db *sql.DB) TransactionRepository {
	return &SQLTransactionRepository{
		db: db,
	}
}

func (r *SQLTransactionRepository) Create(tx *Transaction) (*Transaction, error) {
	return CreateTransaction(r.db, tx)
}

func (r *SQLTransactionRepository) Update(tx Transaction) error {
	return UpdateTransaction(r.db, tx)
}

func (r *SQLTransactionRepository) GetTransactionByGatewayTxnId(gatewayTxnId string) (*Transaction, error) {
	return GetTransactionByGatewayTxnId(r.db, gatewayTxnId)
}
