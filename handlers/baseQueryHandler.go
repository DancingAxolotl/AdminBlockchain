package handlers

import (
	"AdminBlockchain/storage"
	"fmt"
	"os"
	"strings"
)

// BaseQueryHandler a pass-through for acessing the database.
type BaseQueryHandler struct {
	Sp storage.Provider
}

// NewBaseHandler creates a new handler for the specified path
func NewBaseHandler(path string) *BaseQueryHandler {
	var handler BaseQueryHandler
	handler.Load(path)
	return &handler
}

//Load loads the chain state from the specified path
func (handler *BaseQueryHandler) Load(path string) {
	handler.Close()
	handler.Sp.LoadChain(path)

	os.Remove(storage.StateDbPath)
	handler.Sp.StateDb.OpenDb(storage.StateDbPath)

	for _, block := range handler.Sp.Chain {
		handler.AcceptBlock(block)
	}
}

// AcceptBlock at the top of chain
func (handler *BaseQueryHandler) AcceptBlock(block storage.Block) {
	params := strings.Split(block.Data, ";")
	if len(params) > 0 {
		args := make([]interface{}, len(params[1:]))
		for i := range params[1:] {
			args[i] = params[i]
		}
		handler.Sp.StateDb.Transact(params[0], args...)
	}
}

//ExecuteQuery performs a query on the database
func (handler *BaseQueryHandler) ExecuteQuery(query string, params ...interface{}) ([]string, [][]string, error) {
	rows, err := handler.Sp.StateDb.Query(query, params...)
	defer rows.Close()
	if err != nil {
		return nil, nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	vals := make([]interface{}, len(cols))
	raw := make([][]byte, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = &raw[i]
	}

	var rowText [][]string
	for rows.Next() {
		rows.Scan(vals...)
		text := make([]string, len(cols))
		for i, item := range raw {
			text[i] = string(item)
		}
		rowText = append(rowText, text)
	}

	return cols, rowText, nil
}

//ExecuteTransaction performs a transaction and stores it in the blockchain
func (handler *BaseQueryHandler) ExecuteTransaction(query string, params ...interface{}) error {
	err := handler.Sp.StateDb.Transact(query, params...)
	if err != nil {
		return err
	}

	txData := query
	for _, param := range params {
		txData += fmt.Sprintf(";%v", param)
	}

	handler.Sp.Chain.AddBlock(txData)
	return nil
}

//Close saves the state database and closes the connection
func (handler *BaseQueryHandler) Close() {
	handler.Sp.Close()
}
