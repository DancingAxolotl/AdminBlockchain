package rpc

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/network"
)

func main() {
	np := network.NewServerProvider()
	handler := handlers.NewSimpleHandler("./")
	np.RegisterHandler(handler)

	defer np.Stop()

	np.Start("", "8900")
}
