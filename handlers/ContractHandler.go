package handlers

import (
	"AdminBlockchain/utils"
	"errors"
)

const (
	//ContractStatusCreated the contract was just created. No details are final yet.
	ContractStatusCreated = 0
	//ContractStatusConfirmation the contract is being confirmed. No details are final yet.
	ContractStatusConfirmation = 1
	//ContractStatusOpen the contract is open. No details can be changed at this point.
	ContractStatusOpen = 2
	//ContractStatusInProgress work was started on the cotnract.
	ContractStatusInProgress = 3
	//ContractStatusComplete work is complete.
	ContractStatusComplete = 4
	//ContractStatusSuccess work is successful, assignee receives payment.
	ContractStatusSuccess = 5
	//ContractStatusFail work is failed, payment is refunded.
	ContractStatusFail = 6
)

// Balance of tokens of an account
type Balance struct {
	owner   Address
	balance int
}

// Contract information about a contract
type Contract struct {
	ID           int64
	Reporter     Address // who created the contract
	Assignee     Address // who is responsible for performing the task
	ContractInfo string  // off-chain information (e.g. link to specification)
	Status       int     // status of completion
	Reward       int     // reward for completion
}

// ContractHandler handles contract data
type ContractHandler struct {
	*BaseQueryHandler
	Accounts *AccountHandler
}

// Genesis initializes the handler state for new blockchain
func (handler *ContractHandler) Genesis() {
	_, err := handler.ExecuteTransaction(
		"create table Balances (owner text, balance text)")
	utils.LogErrorF(err)

	_, err = handler.ExecuteTransaction(
		"create table Contracts (reporter text, assignee text, contractInfo text, status int, reward int)")
	utils.LogErrorF(err)
}

// GetBalance returns the current user balance
func (handler *ContractHandler) GetBalance(owner Address, createOnErr bool) (int, error) {
	rows, err := handler.Sp.StateDb.Query("select balance from Balances where owner=?", owner)
	if err != nil {
		return -1, err
	}
	var balance int
	if rows.Next() {
		rows.Scan(&balance)
	} else {
		if createOnErr {
			err := handler.setBalance(owner, 100, false)
			return 100, err
		} else {
			return 100, nil
		}
	}
	rows.Close()

	return balance, nil
}

func (handler *ContractHandler) setBalance(owner Address, balance int, update bool) error {
	var err error
	if update {
		_, err = handler.ExecuteTransaction("update Balances set balance=? where owner=?",
			balance,
			owner)
	} else {
		_, err = handler.ExecuteTransaction("insert into Balances (owner, balance) values (?, ?)",
			owner,
			balance)
	}
	return err
}

// CreateContractParams parameters for creating contract
type CreateContractParams struct {
	From         Address // who sends the transaction
	Assignee     Address // who is responsible for performing the task
	ContractInfo string  // off-chain information (e.g. link to specification)
	Reward       int     // reward for the contract
	Signature    []byte
}

// Create creates a contract
func (handler *ContractHandler) Create(params CreateContractParams, contractID *int64) error {
	*contractID = 0
	acc, err := handler.Accounts.getAccountByAddress(params.From)
	if err != nil {
		return err
	}
	balance, err := handler.GetBalance(params.From, true)
	if balance < params.Reward {
		return errors.New("insufficient reporter funds")
	}
	err = checkUserSignature(acc, params.Signature, params.Assignee, params.ContractInfo, params.Reward)
	inserted, err := handler.ExecuteTransaction("insert into Contracts (reporter, assignee, contractInfo, status, reward) values (?, ?, ?, ?, ?)",
		params.From,
		params.Assignee,
		params.ContractInfo,
		ContractStatusCreated,
		params.Reward)
	if err != nil {
		return err
	}

	*contractID = inserted
	return nil
}

// UpdateContractParams parameters for creating contract
type UpdateContractParams struct {
	ContractID   int64
	From         Address // who sends the transaction
	Assignee     Address // who is responsible for performing the task
	ContractInfo string  // off-chain information (e.g. link to specification)
	Reward       int     // reward for the contract
	Signature    []byte
}

// Update updates a contract
func (handler *ContractHandler) Update(params UpdateContractParams, success *bool) error {
	*success = false
	contract, err := handler.getContract(params.ContractID)
	if contract.Status > ContractStatusConfirmation {
		return errors.New("contract already confirmed")
	}
	balance, err := handler.GetBalance(contract.Reporter, true)
	if balance < contract.Reward {
		return errors.New("insufficient reporter funds")
	}
	acc, err := handler.Accounts.getAccountByAddress(params.From)
	if err != nil {
		return err
	}
	err = checkUserSignature(acc, params.Signature, params.ContractID, params.Assignee, params.ContractInfo, params.Reward)
	_, err = handler.ExecuteTransaction("update Contracts set assignee=?, contractInfo=?, status=?, reward=? where rowid=?",
		params.Assignee,
		params.ContractInfo,
		ContractStatusCreated,
		params.Reward,
		params.ContractID)
	if err != nil {
		return err
	}

	*success = true
	return nil
}

// ContractStateParams parameters to update contract
type ContractStateParams struct {
	ContractID int64
	From       Address // who sends the transaction
	Signature  []byte
}

// Sign signs the contract
func (handler *ContractHandler) Sign(params UpdateContractParams, success *bool) error {
	*success = false
	contract, err := handler.getContract(params.ContractID)
	if contract.Status > ContractStatusConfirmation {
		return errors.New("contract already confirmed")
	}
	balanceR, err := handler.GetBalance(contract.Reporter, true)
	if balanceR < contract.Reward {
		return errors.New("insufficient reporter funds")
	}
	acc, err := handler.Accounts.getAccountByAddress(contract.Assignee)
	if err != nil {
		return err
	}

	err = checkUserSignature(acc, params.Signature, params.ContractID)
	if err != nil {
		return err
	}

	err = handler.updateStatus(params.ContractID, ContractStatusOpen)
	if err != nil {
		return err
	}

	err = handler.setBalance(contract.Reporter, balanceR-contract.Reward, true)
	if err != nil {
		return err
	}

	*success = true
	return nil
}

// StartProgress start progress on contract
func (handler *ContractHandler) StartProgress(params UpdateContractParams, success *bool) error {
	*success = false
	contract, err := handler.getContract(params.ContractID)
	if contract.Status != ContractStatusOpen {
		return errors.New("contract is not available")
	}
	acc, err := handler.Accounts.getAccountByAddress(contract.Assignee)
	if err != nil {
		return err
	}

	err = checkUserSignature(acc, params.Signature, params.ContractID)
	if err != nil {
		return err
	}

	err = handler.updateStatus(params.ContractID, ContractStatusInProgress)
	if err != nil {
		return err
	}

	*success = true
	return nil
}

// Resolve finish work on contract
func (handler *ContractHandler) Resolve(params UpdateContractParams, success *bool) error {
	*success = false
	contract, err := handler.getContract(params.ContractID)
	if contract.Status != ContractStatusInProgress {
		return errors.New("contract is not available")
	}
	acc, err := handler.Accounts.getAccountByAddress(contract.Assignee)
	if err != nil {
		return err
	}

	err = checkUserSignature(acc, params.Signature, params.ContractID)
	if err != nil {
		return err
	}

	err = handler.updateStatus(params.ContractID, ContractStatusComplete)
	if err != nil {
		return err
	}

	*success = true
	return nil
}

// ContractAcceptanceParams parameters to update contract
type ContractAcceptanceParams struct {
	ContractID int64
	From       Address // who sends the transaction
	Success    bool    // is acceptance succesful
	Signature  []byte
}

// Acceptance accept the completed work
func (handler *ContractHandler) Acceptance(params ContractAcceptanceParams, success *bool) error {
	*success = false
	contract, err := handler.getContract(params.ContractID)
	if contract.Status != ContractStatusComplete {
		return errors.New("contract is not complete")
	}
	acc, err := handler.Accounts.getAccountByAddress(contract.Reporter)
	if err != nil {
		return err
	}

	err = checkUserSignature(acc, params.Signature, params.ContractID, params.Success)
	if err != nil {
		return err
	}

	balanceR, err := handler.GetBalance(contract.Reporter, true)
	balanceA, err := handler.GetBalance(contract.Assignee, true)
	if err != nil {
		return err
	}

	if params.Success {
		err = handler.updateStatus(params.ContractID, ContractStatusSuccess)
		if err != nil {
			return err
		}
		handler.setBalance(contract.Assignee, balanceA+contract.Reward, true)
	} else {
		err = handler.updateStatus(params.ContractID, ContractStatusFail)
		if err != nil {
			return err
		}
		handler.setBalance(contract.Reporter, balanceR+contract.Reward, true)
	}

	*success = true
	return nil
}

// GetAllContracts retruns the list of contracts
func (handler *ContractHandler) GetAllContracts() ([]Contract, error) {
	var contracts []Contract
	rows, err := handler.Sp.StateDb.Query("select rowid, reporter, assignee, contractInfo, status, reward from Contracts")
	defer rows.Close()
	if err != nil {
		return []Contract{}, err
	}

	var contract Contract
	for rows.Next() {
		rows.Scan(&contract.ID, &contract.Reporter, &contract.Assignee, &contract.ContractInfo, &contract.Status, &contract.Reward)
		contracts = append(contracts, contract)
	}

	return contracts, nil
}

// GetContractsOfUser retruns the list of contracts of user
func (handler *ContractHandler) GetContractsOfUser(user Address) ([]Contract, error) {
	var contracts []Contract
	rows, err := handler.Sp.StateDb.Query("select rowid, reporter, assignee, contractInfo, status, reward from Contracts where assignee=? or reporter=?", user, user)
	defer rows.Close()
	if err != nil {
		return []Contract{}, err
	}

	var contract Contract
	for rows.Next() {
		rows.Scan(&contract.ID, &contract.Reporter, &contract.Assignee, &contract.ContractInfo, &contract.Status, &contract.Reward)
		contracts = append(contracts, contract)
	}

	return contracts, nil
}

func (handler *ContractHandler) getContract(id int64) (Contract, error) {
	var contract Contract
	rows, err := handler.Sp.StateDb.Query("select rowid, reporter, assignee, contractInfo, status, reward from Contracts where rowid=?", id)
	defer rows.Close()
	if err != nil {
		return contract, err
	}

	if rows.Next() {
		rows.Scan(&contract.ID, &contract.Reporter, &contract.Assignee, &contract.ContractInfo, &contract.Status, &contract.Reward)
	}

	return contract, nil
}

func (handler *ContractHandler) updateStatus(id int64, status int) error {
	_, err := handler.ExecuteTransaction("update Contracts set status=? where rowid=?",
		status,
		id)
	return err
}

func checkUserSignature(acc Account, signature []byte, params ...interface{}) error {
	err := acc.PubKey.CheckSignature(
		utils.Hash(params...),
		signature)
	if err != nil {
		return errors.New("invalid user signature")
	}

	return nil
}
