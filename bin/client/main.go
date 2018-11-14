package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/handlers/sample"
	"AdminBlockchain/utils"
	//"bufio"
	//"os"
	"encoding/hex"
	"fmt"
	"log"
	"net/rpc"
	"strings"
	"time"
)

var (
	clientKey      utils.SignatureCreator
	clientAddress  handlers.Address
	accountHandler handlers.AccountHandler
	client         *rpc.Client
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
			time.Sleep(5 * time.Second)
		}
	}
}

func stopSync(stop chan bool) {
	stop <- true
}

func main() {
	// Prepare rpc connection
	log.Print("Connecting...")
	var err error
	client, err = rpc.DialHTTP("tcp", "localhost:8900")
	utils.LogErrorF(err)
	defer client.Close()

	// Create base handler for transactions
	baseHandler := handlers.NewBaseHandler("./")
	defer baseHandler.Close()
	// Define handlers
	accountHandler = handlers.AccountHandler{BaseQueryHandler: baseHandler}

	// Set up block synchronization
	serverKey, err := utils.LoadPublicKey("./server.pem")
	utils.LogErrorF(err)
	blockSync := handlers.BlockSyncHandler{StorageProvider: &baseHandler.Sp, QueryHandlers: []handlers.IHandler{accountHandler}, SignValidator: serverKey}
	blockProvider := handlers.RPCBlockProvider{Client: client}

	syncChan := make(chan bool)
	go syncClient(&blockSync, &blockProvider, syncChan)
	defer stopSync(syncChan)

	clientKey, err = utils.LoadPrivateKey("./private.pem")
	utils.LogErrorF(err)
	clientAddress = utils.Hash(utils.LoadPublicKey("./public.pem"))

	// Start input loop
	//reader := bufio.NewReader(os.Stdin)
	fmt.Print("Available commands: accounts, state, help, exit\n")
	var running = true
	for running {
		fmt.Print("> ")
		//input, _ := reader.ReadString('\n')
		input := "accounts get"
		var command string
		fmt.Sscan(input, &command)
		switch command {
		case "help":
			fmt.Print(
				"Commands:\n" +
					"  accounts - manage user accounts\n" +
					"    get - list all current accounts (local)\n" +
					"    create <path to public key> <personal info> <access level> - create a new user account. By default access level is basic.\n" +
					"    update <address> <personal info> <access level> - update data about user accountn" +
					"  state - prints the current blockchain state. (local)\n" +
					"  help - prints this help message.\n" +
					"  exit - exits.\n")

		case "state":
			fmt.Print(" Block id      | Previous hash    | Block hash       | Data \n")
			for _, item := range baseHandler.Sp.Chain {
				fmt.Printf(" %-14d|%14.14s ...|%14.14s ...| %.80s\n",
					item.ID,
					fmt.Sprintf("% x", item.PrevHash),
					fmt.Sprintf("% x", item.Hash()),
					item.Data)
			}

		case "accounts":
			handleAccounts(input)

		case "exit":
			running = false
		}

	}
}

func handleAccounts(input string) {
	var command string
	fmt.Sscanf(input, "accounts %s", &command)
	switch command {
	case "get":
		accounts := accountHandler.ListAccounts()
		fmt.Printf(" Address        | Personal info    | Access level\n")
		for _, account := range accounts {
			var accessLvl, tmp string
			if account.AccessLevel == handlers.BasicAccountAccess {
				accessLvl = "basic"
			} else {
				accessLvl = "admin"
			}
			tmp = account.PersonalInfo
			fmt.Printf(" %14.14s | %s | %s\n",
				fmt.Sprintf("%x", account.Address),
				tmp,
				accessLvl)
		}
	case "create":
		var pubKeyPath, personalInfo string
		var access int
		fmt.Sscanf(input, "accounts create %q %q %d", &pubKeyPath, &personalInfo, &access)
		publicKey, err := utils.LoadPublicKey(pubKeyPath)
		utils.LogErrorF(err)
		signature, err := clientKey.Sign(utils.Hash(personalInfo, access, publicKey))
		utils.LogErrorF(err)

		err = client.Call("AccountHandler.CreateAccount", handlers.CreateAccountParams{
			From:         clientAddress,
			PersonalInfo: personalInfo,
			AccessLevel:  access,
			PubKey:       publicKey,
			Signature:    signature}, nil)
		utils.LogErrorF(err)

	case "update":
		var addressStr, personalInfo string
		var access int
		fmt.Sscanf(input, "accounts update %q %q %d", &addressStr, &personalInfo, &access)
		address, err := hex.DecodeString(addressStr)
		utils.LogErrorF(err)
		signature, err := clientKey.Sign(utils.Hash(address, personalInfo, access))
		utils.LogErrorF(err)

		err = client.Call("AccountHandler.UpdateAccount", handlers.UpdateAccountParams{
			From:         clientAddress,
			Account:      address,
			PersonalInfo: personalInfo,
			AccessLevel:  access,
			Signature:    signature}, nil)
		utils.LogErrorF(err)
	}
}

func printTable(resp sample.SimpleHandlerResponce) {
	fmt.Printf("%v\n", strings.Join(resp.Columns, "\t|"))
	for _, row := range resp.Rows {
		fmt.Printf("%v\n", strings.Join(row, "\t|"))
	}
}
