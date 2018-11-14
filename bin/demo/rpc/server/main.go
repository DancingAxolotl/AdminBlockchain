package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/handlers/sample"
	"AdminBlockchain/network"
	"AdminBlockchain/utils"
)

func main() {
	np := network.NewServerProvider()
	defer np.Stop()

	handler := sample.NewSimpleHandler("./")
	key, err := utils.LoadPrivateKey("./private.pem")
	utils.LogErrorF(err)
	var blockHandler = handlers.BlockPropagationHandler{Storage: &handler.Sp, Signer: key}
	defer handler.Close()

	np.RegisterHandler(handler)
	np.RegisterHandler(&blockHandler)

	np.Start("", "8900")
}
