package handlers

import (
	"AdminBlockchain/storage"
	"fmt"
)

// SimpleQueryHandler a pass-through for acessing the database. Provides simple logic for storing each executed transaction on the blockchain.
type SimpleQueryHandler struct {
	Sp storage.StorageProvider
}

// SimpleHandlerRequest request parameters for the SimpleQueryHandler
type SimpleHandlerRequest struct {
	query  string
	params []interface{}
}

// SimpleHandlerResponce responce from the SimpleQueryHandler
type SimpleHandlerResponce struct {
	columns []string
	rows    [][]string
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
func (handler *SimpleQueryHandler) ExecuteQuery(request SimpleHandlerRequest, responce *SimpleHandlerResponce) error {
	rows, err := handler.Sp.StateDb.Query(request.query)
	if err != nil {
		return err
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	vals := make([]interface{}, len(cols))
	raw := make([][]byte, len(cols))
	(*responce).columns = make([]string, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = &raw[i]
		(*responce).columns[i] = cols[i]
	}

	for rows.Next() {
		rows.Scan(vals...)

		text := make([]string, len(cols))
		for i, item := range raw {
			text[i] = string(item)
		}

		(*responce).rows = append((*responce).rows, text)
	}

	return nil
}

//ExecuteTransaction performs a transaction and stores it in the blockchain
func (handler *SimpleQueryHandler) ExecuteTransaction(request SimpleHandlerRequest, responce *bool) error {
	*responce = false
	err := handler.Sp.StateDb.Transact(request.query, request.params...)
	if err != nil {
		return err
	}

	txData := request.query
	for _, param := range request.params {
		txData += fmt.Sprintf(";%v", param)
	}

	handler.Sp.Chain.AddBlock(txData)
	*responce = true
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
