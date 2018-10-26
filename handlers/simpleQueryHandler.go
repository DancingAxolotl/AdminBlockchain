package handlers

import (
	"AdminBlockchain/storage"
	"database/sql"
	"fmt"
)

//SimpleQueryHandler a pass-through for acessing the database. Provides simple logic for storing each executed transaction on the blockchain.
type SimpleQueryHandler struct {
	Sp storage.StorageProvider
}

// NewHandler creates a new handler for the specified path
func NewHandler(path string) *SimpleQueryHandler {
	var handler SimpleQueryHandler
	handler.Load(path)
	return &handler
}

//Load loads the chain state from the specified path
func (handler *SimpleQueryHandler) Load(path string) {
	handler.Close()
	handler.Sp.LoadChain(path)
}

//ExecuteQuery performs a query on the database
func (handler *SimpleQueryHandler) ExecuteQuery(query string) (*sql.Rows, error) {
	return handler.Sp.StateDb.Query(query)
}

//ExecuteTransaction performs a transaction and stores it in the blockchain
func (handler *SimpleQueryHandler) ExecuteTransaction(statement string, params ...interface{}) error {
	err := handler.Sp.StateDb.Transact(statement, params...)
	if err != nil {
		return err
	}

	txData := statement
	for _, param := range params {
		txData += fmt.Sprintf("%v", param)
	}

	handler.Sp.Chain.AddBlock(txData)

	return nil
}

//Close saves the state database and closes the connection
func (handler *SimpleQueryHandler) Close() {
	if handler.Sp.ChainDb.IsOpen() {
		handler.Sp.UpdateChainState()
		handler.Sp.ChainDb.Close()
		handler.Sp.StateDb.Close()
	}
}
