package local

import (
	"AdminBlockchain/handlers"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	handler := handlers.NewSimpleHandler("./")
	defer handler.Close()

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
					"  query \"<sql select query>\" - sends a query to the database, prints the output.\n" +
					"  execute \"<sql statement>\" - executes an sql transaction that changes the database.\n" +
					"  state - prints the current blockchain state.\n" +
					"  help - prints this help message.\n" +
					"  exit - exits.\n")

		case "state":
			fmt.Print(" Block id      | Previous hash    | Block hash       | Data \n")
			for _, item := range handler.Sp.Chain {
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
			err := handler.ExecuteQuery(handlers.SimpleHandlerRequest{Query: query, Params: []interface{}{}}, &resp)
			if err == nil {
				printTable(resp)
			} else {
				log.Fatal(err)
			}

		case "execute":
			var query string
			fmt.Sscanf(input, "execute%q", &query)
			var success bool
			err := handler.ExecuteTransaction(handlers.SimpleHandlerRequest{Query: query, Params: []interface{}{}}, &success)
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
