package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/network"
)

func main() {
	np := network.NewServerProvider()
	defer np.Stop()

	handler := handlers.NewSimpleHandler("./")
	var blockHandler = handlers.BlockPropagationHandler{Storage: &handler.Sp}
	defer handler.Close()

	np.RegisterHandler(handler)
	np.RegisterHandler(&blockHandler)

	np.Start("", "8900")
}
