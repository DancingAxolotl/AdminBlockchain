package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/network"
	"log"
)

type TestHandler int
type Stub int

func (th *TestHandler) Test(request Stub, responce *Stub) error {
	log.Print("Test method called.")
	return nil
}

func main() {
	np := network.NewServerProvider()
	defer np.Stop()

	handler := handlers.NewSimpleHandler("./")
	defer handler.Close()

	np.RegisterHandler(handler)
	var th TestHandler
	np.RegisterHandler(&th)

	np.Start("", "8900")
}
