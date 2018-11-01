package main

import (
	"log"
	"net/rpc"
)

func main() {
	log.Print("Start connection")
	client, err := rpc.DialHTTP("tcp", "localhost:8900")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	var temp int
	client.Call("TestHandler.Test", 1, &temp)
	client.Close()
}
