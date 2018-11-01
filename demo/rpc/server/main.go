package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/network"
	"bufio"
	"fmt"
	"log"
	"os"
)

type TestHandler int
type Stub int

func (th *TestHandler) Test(request Stub, responce *Stub) error {
	log.Print("Test method called.")
	return nil
}

func start(np *network.ServerNetworkProvider) {
	np.Start("", "8900")
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	np := network.NewServerProvider()
	handler := handlers.NewSimpleHandler("./")
	np.RegisterHandler(handler)
	var th TestHandler
	np.RegisterHandler(&th)

	fmt.Print("Available commands: start, stop, exit\n")
	var running = true
	for running {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		var command string
		fmt.Sscan(input, &command)
		switch command {
		case "start":
			go start(&np)
		case "stop":
			np.Stop()
		case "exit":
			running = false
		}
	}
}
