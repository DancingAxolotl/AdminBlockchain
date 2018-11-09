package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/utils"
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"time"
)

func syncClient(sync *handlers.BlockSyncHandler, rpc *handlers.RPCBlockProvider, stop chan bool) {
	for {
		select {
		case <-stop:
			return
		default:
			err := sync.Sync(rpc)
			if err != nil {
				log.Print(err)
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func stopSync(stop chan bool) {
	stop <- true
}

func main() {
	log.Print("Connecting...")
	client, err := rpc.DialHTTP("tcp", "localhost:8900")
	utils.LogErrorF(err)

	defer client.Close()

	reader := bufio.NewReader(os.Stdin)
	localChainHandler := handlers.NewSimpleHandler("./")
	defer localChainHandler.Close()

	blockSync := handlers.BlockSyncHandler{QueryHandler: localChainHandler}
	blockProvider := handlers.RPCBlockProvider{Client: client}
	syncChan := make(chan bool)
	go syncClient(&blockSync, &blockProvider, syncChan)
	defer stopSync(syncChan)

	fmt.Print("Available commands: query, execute, state, help, exit\n")
	var running = true
	for running {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		var command string
		fmt.Sscan(input, &command)
		switch command {
		case "help":
			fmt.Print(
				"Commands:\n" +
					"  query \"<sql select query>\" - sends a query to the database, prints the output. (local)\n" +
					"  execute \"<sql statement>\" - executes an sql transaction that changes the database. (send to server)\n" +
					"  state - prints the current blockchain state. (local)\n" +
					"  help - prints this help message.\n" +
					"  exit - exits.\n")

		case "state":
			fmt.Print(" Block id      | Previous hash    | Block hash       | Data \n")
			for _, item := range localChainHandler.Sp.Chain {
				fmt.Printf(" %-14d|%14.14s ...|%14.14s ...| %v\n",
					item.ID,
					fmt.Sprintf("% x", item.PrevHash),
					fmt.Sprintf("% x", item.Hash()),
					item.Data)
			}

		case "query":
			var query string
			fmt.Sscanf(input, "query%q", &query)
			var resp handlers.SimpleHandlerResponce
			err := localChainHandler.ExecuteQuery(handlers.SimpleHandlerRequest{Query: query, Params: []interface{}{}}, &resp)
			if err == nil {
				printTable(resp)
			} else {
				log.Fatal(err)
			}

		case "execute":
			var query string
			fmt.Sscanf(input, "execute%q", &query)
			var success bool
			err := client.Call("SimpleQueryHandler.ExecuteTransaction", handlers.SimpleHandlerRequest{Query: query, Params: []interface{}{}}, &success)
			if err != nil {
				log.Fatal(err)
			}

		case "exit":
			running = false
		}

	}

}

func printTable(resp handlers.SimpleHandlerResponce) {
	fmt.Printf("%v\n", strings.Join(resp.Columns, "\t|"))
	for _, row := range resp.Rows {
		fmt.Printf("%v\n", strings.Join(row, "\t|"))
	}
}
