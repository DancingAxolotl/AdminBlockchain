package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/network"
	"AdminBlockchain/utils"
	"os"
	"os/signal"
)

var (
	np          network.ServerNetworkProvider
	baseHandler *handlers.BaseQueryHandler
)

func handleStop() {
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan

	np.Stop()
	baseHandler.Close()

	os.Exit(0)
}

func main() {
	np = network.NewServerProvider()
	baseHandler = handlers.NewBaseHandler("./")

	key, err := utils.LoadPrivateKey("./private.pem")
	utils.LogErrorF(err)
	var blockHandler = handlers.BlockPropagationHandler{Storage: &baseHandler.Sp, Signer: key}

	accHandler := handlers.AccountHandler{BaseQueryHandler: baseHandler}
	contractHandler := handlers.ContractHandler{BaseQueryHandler: baseHandler, Accounts: &accHandler}

	if len(baseHandler.Sp.Chain) == 0 {
		key, err := utils.LoadPublicKey("./public.pem")
		utils.LogErrorF(err)
		accHandler.Genesis(key)
		contractHandler.Genesis()
	}

	np.RegisterHandler(&accHandler)
	np.RegisterHandler(&contractHandler)
	np.RegisterHandler(&blockHandler)
	go handleStop()
	np.Start("", "8900")
}
