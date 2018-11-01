package main

import (
	"AdminBlockchain/handlers"
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	handler := handlers.NewHandler("./")
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
					item.Id,
					fmt.Sprintf("% x", item.PrevHash),
					fmt.Sprintf("% x", item.Hash()),
					item.Data)
			}

		case "query":
			var query string
			fmt.Sscanf(input, "query%q", &query)
			rows, err := handler.ExecuteQuery(query)
			if err == nil {
				printTable(rows)
			} else {
				log.Fatal(err)
			}
			rows.Close()

		case "execute":
			var query string
			fmt.Sscanf(input, "execute%q", &query)
			err := handler.ExecuteTransaction(query)
			if err != nil {
				log.Fatal(err)
			}

		case "exit":
			running = false
		}

	}

}

func printTable(rows *sql.Rows) {
	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", strings.Join(cols, "\t|"))
	vals := make([]interface{}, len(cols))
	raw := make([][]byte, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = &raw[i]
	}

	for rows.Next() {
		rows.Scan(vals...)
		text := make([]string, len(cols))
		for i, item := range raw {
			text[i] = string(item)
		}
		fmt.Printf("%v\n", strings.Join(text, "\t|"))
	}
}
