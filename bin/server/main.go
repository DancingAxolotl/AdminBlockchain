package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/network"
	"AdminBlockchain/utils"
)

func main() {
	np := network.NewServerProvider()
	defer np.Stop()
	baseHandler := handlers.NewBaseHandler("./")
	defer baseHandler.Close()

	key, err := utils.LoadPrivateKey("./private.pem")
	utils.LogErrorF(err)
	var blockHandler = handlers.BlockPropagationHandler{Storage: &baseHandler.Sp, Signer: key}

	accHandler := handlers.AccountHandler{BaseQueryHandler: baseHandler}

	if len(baseHandler.Sp.Chain) == 0 {
		key, err := utils.LoadPublicKey("./public.pem")
		utils.LogErrorF(err)
		accHandler.Genesis(key)
	}

	np.RegisterHandler(&accHandler)
	np.RegisterHandler(&blockHandler)

	np.Start("", "8900")
}
