package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/network"
	"AdminBlockchain/utils"
	"time"
)

func updateServerChainState(handler *handlers.BaseQueryHandler, stop chan bool) {
	for {
		select {
		case <-stop:
			return
		default:
			handler.Sp.UpdateChainState()
			time.Sleep(5 * time.Second)
		}
	}
}

func stopUpdate(stop chan bool) {
	stop <- true
}

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

	updChan := make(chan bool)
	go updateServerChainState(baseHandler, updChan)
	defer stopUpdate(updChan)

	np.RegisterHandler(&accHandler)
	np.RegisterHandler(&blockHandler)

	np.Start("", "8900")
}
