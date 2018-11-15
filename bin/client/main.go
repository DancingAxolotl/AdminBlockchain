package main

import (
	"AdminBlockchain/handlers"
	"AdminBlockchain/utils"
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"time"
)

var (
	clientKey       utils.SignatureCreator
	clientAddress   handlers.Address
	accountHandler  handlers.AccountHandler
	contractHandler handlers.ContractHandler
	client          *rpc.Client
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
	contractHandler = handlers.ContractHandler{BaseQueryHandler: baseHandler, Accounts: &accountHandler}

	// Set up block synchronization
	serverKey, err := utils.LoadPublicKey("./server.pem")
	utils.LogErrorF(err)
	blockSync := handlers.BlockSyncHandler{StorageProvider: &baseHandler.Sp, QueryHandlers: []handlers.IHandler{accountHandler}, SignValidator: serverKey}
	blockProvider := handlers.RPCBlockProvider{Client: client}

	syncChan := make(chan bool)
	go syncClient(&blockSync, &blockProvider, syncChan)
	defer stopSync(syncChan)

	// Load client keys
	clientKey, err = utils.LoadPrivateKey("./private.pem")
	utils.LogErrorF(err)
	tmpKey, err := utils.LoadPublicKey("./public.pem")
	utils.LogErrorF(err)
	clientAddress = handlers.GetAddressFromPubKey(tmpKey)

	// Start input loop
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Available commands: accounts, contracts, balance, state, help, exit\n")
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
					"  accounts - manage user accounts\n" +
					"    get - list all current accounts (local)\n" +
					"    create <path to public key> <personal info> <access level> - create a new user account. By default access level is basic.\n" +
					"    update <address> <personal info> <access level> - update data about user accountn\n" +
					"  contracts - manage contracts\n" +
					"    get - list all contracts\n" +
					"    getmy - list user contracts\n" +
					"    create <assignee> <info> <reward> - create a new contract\n" +
					"    update <id> <assignee> <info> <reward> - create a new contract\n" +
					"    sign <id> - sign the contract\n" +
					"    start <id> - start progress on the contract\n" +
					"    resolve <id> - resolve the contract\n" +
					"    accept <id> <accepted> - acceptance of the contract\n" +
					"  balance - prints users balance\n" +
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

		case "contracts":
			handleContracts(input)

		case "balance":
			balance, err := contractHandler.GetBalance(clientAddress, false)
			utils.LogError(err)
			fmt.Printf("Balance of user %v is %v\n", clientAddress, balance)

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
			var accessLvl string
			if account.AccessLevel == handlers.BasicAccountAccess {
				accessLvl = "basic"
			} else {
				accessLvl = "admin"
			}
			fmt.Printf(" %14.14s | %16.16s | %11.11s\n",
				account.Address,
				account.PersonalInfo,
				accessLvl)
		}
	case "create":
		var pubKeyPath, personalInfo string
		var access int
		fmt.Sscanf(input, "accounts create %q %q %d", &pubKeyPath, &personalInfo, &access)
		publicKey, err := utils.LoadPublicKey(pubKeyPath)
		utils.LogErrorF(err)
		pubKeyData, err := publicKey.Store()
		utils.LogErrorF(err)
		signature, err := clientKey.Sign(utils.Hash(personalInfo, access, pubKeyData))
		utils.LogErrorF(err)

		err = client.Call("AccountHandler.CreateAccount", handlers.CreateAccountParams{
			From:         clientAddress,
			PersonalInfo: personalInfo,
			AccessLevel:  access,
			PubKey:       pubKeyData,
			Signature:    signature}, nil)
		utils.LogErrorF(err)

	case "update":
		var addressStr, personalInfo string
		var access int
		fmt.Sscanf(input, "accounts update %s %q %d", &addressStr, &personalInfo, &access)
		signature, err := clientKey.Sign(utils.Hash(addressStr, personalInfo, access))
		utils.LogErrorF(err)

		err = client.Call("AccountHandler.UpdateAccount", handlers.UpdateAccountParams{
			From:         clientAddress,
			Account:      handlers.Address(addressStr),
			PersonalInfo: personalInfo,
			AccessLevel:  access,
			Signature:    signature}, nil)
		utils.LogErrorF(err)
	}
}

func handleContracts(input string) {
	var command string
	fmt.Sscanf(input, "contracts %s", &command)
	switch command {
	case "get":
		contracts, err := contractHandler.GetAllContracts()
		utils.LogError(err)
		printContracts(contracts)
	case "getmy":
		contracts, err := contractHandler.GetContractsOfUser(clientAddress)
		utils.LogError(err)
		printContracts(contracts)
	case "create":
		var Assignee, ContractInfo string
		var Reward int
		fmt.Sscanf(input, "contracts create %q %q %d", &Assignee, &ContractInfo, &Reward)
		signature, err := clientKey.Sign(utils.Hash(Assignee, ContractInfo, Reward))
		utils.LogErrorF(err)
		var contractID int64
		err = client.Call("ContractHandler.Create", handlers.CreateContractParams{
			From:         clientAddress,
			Assignee:     handlers.Address(Assignee),
			ContractInfo: ContractInfo,
			Reward:       Reward,
			Signature:    signature}, &contractID)
		utils.LogError(err)
		if err == nil {
			fmt.Printf("Contract created - %v", contractID)
		}

	case "update":
		var Assignee, ContractInfo string
		var Reward int
		var ID int64
		fmt.Sscanf(input, "contracts update %d %q %q %d", &ID, &Assignee, &ContractInfo, &Reward)
		signature, err := clientKey.Sign(utils.Hash(ID, Assignee, ContractInfo, Reward))
		utils.LogErrorF(err)
		var tmp bool
		err = client.Call("ContractHandler.Update", handlers.UpdateContractParams{
			ContractID:   ID,
			From:         clientAddress,
			Assignee:     handlers.Address(Assignee),
			ContractInfo: ContractInfo,
			Reward:       Reward,
			Signature:    signature}, &tmp)
		utils.LogError(err)

	case "sign":
		var ID int64
		fmt.Sscanf(input, "contracts sign %d", &ID)
		signature, err := clientKey.Sign(utils.Hash(ID))
		utils.LogErrorF(err)
		var tmp bool
		err = client.Call("ContractHandler.Sign", handlers.ContractStateParams{
			ContractID: ID,
			From:       clientAddress,
			Signature:  signature}, &tmp)
		utils.LogError(err)

	case "start":
		var ID int64
		fmt.Sscanf(input, "contracts start %d", &ID)
		signature, err := clientKey.Sign(utils.Hash(ID))
		utils.LogErrorF(err)
		var tmp bool
		err = client.Call("ContractHandler.StartProgress", handlers.ContractStateParams{
			ContractID: ID,
			From:       clientAddress,
			Signature:  signature}, &tmp)
		utils.LogError(err)

	case "resolve":
		var ID int64
		fmt.Sscanf(input, "contracts resolve %d", &ID)
		signature, err := clientKey.Sign(utils.Hash(ID))
		utils.LogErrorF(err)
		var tmp bool
		err = client.Call("ContractHandler.Resolve", handlers.ContractStateParams{
			ContractID: ID,
			From:       clientAddress,
			Signature:  signature}, &tmp)
		utils.LogError(err)

	case "accept":
		var ID int64
		var success bool
		fmt.Sscanf(input, "contracts accept %d %t", &ID, &success)
		signature, err := clientKey.Sign(utils.Hash(ID, success))
		utils.LogErrorF(err)
		var tmp bool
		err = client.Call("ContractHandler.Acceptance", handlers.ContractAcceptanceParams{
			ContractID: ID,
			From:       clientAddress,
			Success:    success,
			Signature:  signature}, &tmp)
		utils.LogError(err)
	}
}

func printContracts(contracts []handlers.Contract) {
	fmt.Printf(" ID | Reporter       | Assignee         | Info             | Status | Reward \n")
	for _, item := range contracts {
		fmt.Printf(" %2.d | %14.14s | %16.16s | %16.16s | %6.d | %d \n",
			item.ID,
			item.Reporter,
			item.Assignee,
			item.ContractInfo,
			item.Status,
			item.Reward)
	}
}
