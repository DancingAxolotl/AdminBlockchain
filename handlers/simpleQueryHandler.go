package handlers

import (
	"AdminBlockchain/storage"
	"fmt"
)

// SimpleQueryHandler a pass-through for acessing the database. Provides simple logic for storing each executed transaction on the blockchain.
type SimpleQueryHandler struct {
	Sp storage.Provider
}

// SimpleHandlerRequest request parameters for the SimpleQueryHandler
type SimpleHandlerRequest struct {
	Query  string
	Params []interface{}
}

// SimpleHandlerResponce responce from the SimpleQueryHandler
type SimpleHandlerResponce struct {
	Columns []string
	Rows    [][]string
}

// NewHandler creates a new handler for the specified path
func NewSimpleHandler(path string) *SimpleQueryHandler {
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
	rows, err := handler.Sp.StateDb.Query(request.Query)
	defer rows.Close()
	if err != nil {
		return err
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	vals := make([]interface{}, len(cols))
	raw := make([][]byte, len(cols))
	(*responce).Columns = make([]string, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = &raw[i]
		(*responce).Columns[i] = cols[i]
	}

	for rows.Next() {
		rows.Scan(vals...)

		text := make([]string, len(cols))
		for i, item := range raw {
			text[i] = string(item)
		}

		(*responce).Rows = append((*responce).Rows, text)
	}

	return nil
}

//ExecuteTransaction performs a transaction and stores it in the blockchain
func (handler *SimpleQueryHandler) ExecuteTransaction(request SimpleHandlerRequest, responce *bool) error {
	*responce = false
	err := handler.Sp.StateDb.Transact(request.Query, request.Params...)
	if err != nil {
		return err
	}

	txData := request.Query
	for _, param := range request.Params {
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
