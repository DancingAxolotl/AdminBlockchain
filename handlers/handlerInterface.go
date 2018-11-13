package handlers

import (
	"AdminBlockchain/storage"
)

//IHandler interface for rpc handlers
type IHandler interface {
	Init()
	AcceptBlock(storage.Block)
}
