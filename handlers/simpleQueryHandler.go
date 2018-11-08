package handlers

import (
	"AdminBlockchain/storage"
	"fmt"
	"log"
	"os"
	"strings"
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

// NewSimpleHandler creates a new handler for the specified path
func NewSimpleHandler(path string) *SimpleQueryHandler {
	var handler SimpleQueryHandler
	handler.Load(path)
	return &handler
}

var stateDbPath = "./storage.db"

//Load loads the chain state from the specified path
func (handler *SimpleQueryHandler) Load(path string) {
	handler.Close()
	handler.Sp.LoadChain(path)

	os.Remove(stateDbPath)
	handler.Sp.StateDb.OpenDb(stateDbPath)

	for _, block := range handler.Sp.Chain {
		handler.AcceptBlock(block)
	}
}

// AcceptBlock at the top of chain
func (handler *SimpleQueryHandler) AcceptBlock(block storage.Block) {
	params := strings.Split(block.Data, ";")
	if len(params) > 1 {
		args := make([]interface{}, len(params[1:]))
		for i := range params[1:] {
			args[i] = params[i]
		}
		handler.Sp.StateDb.Transact(params[0], args...)
	}
}

//ExecuteQuery performs a query on the database
func (handler *SimpleQueryHandler) ExecuteQuery(request SimpleHandlerRequest, responce *SimpleHandlerResponce) error {
	log.Printf("ExecuteQuery called with request: %v", request.Query)
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
	log.Printf("ExecuteTransaction called with request %v", request.Query)
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
		log.Print("Handler closing...")
		handler.Sp.UpdateChainState()
		handler.Sp.ChainDb.Close()
		handler.Sp.StateDb.Close()
	}
}
