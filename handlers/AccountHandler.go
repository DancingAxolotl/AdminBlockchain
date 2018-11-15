package handlers

import (
	"AdminBlockchain/utils"
	"errors"
	"fmt"
)

// AccessLevels
const (
	BasicAccountAccess = 0
	AdminAccountAccess = 1
)

// Address an address of an account
type Address string

// Account basic account info
type Account struct {
	Address      Address                  // derived from public key
	PersonalInfo string                   // Information to identify the account owner
	AccessLevel  int                      // Account access level
	PubKey       utils.SignatureValidator // To validate user signature
}

// AccountHandler handles account data
type AccountHandler struct {
	*BaseQueryHandler
}

// Genesis initializes the handler state for new blockchain
func (handler *AccountHandler) Genesis(PublicKey utils.SignatureValidator) {
	handler.ExecuteTransaction(
		"create table Accounts (address text, personal text, level int, pkey blob)")

	accAddress := GetAddressFromPubKey(PublicKey)
	key, err := PublicKey.Store()
	utils.LogErrorF(err)

	_, err = handler.ExecuteTransaction(
		"insert into Accounts (address, personal, level, pkey) values (?, ?, ?, ?)", accAddress, "admin", AdminAccountAccess, key)
	utils.LogErrorF(err)
}

// CreateAccountParams for updating or creating an account
type CreateAccountParams struct {
	From         Address // who adds the account
	PersonalInfo string  // personal info of the new account
	AccessLevel  int     // access level of the new account
	PubKey       []byte  // public key of the new account
	Signature    []byte  // sender signature
}

// CreateAccount creates an account
func (handler *AccountHandler) CreateAccount(params CreateAccountParams, sucess *bool) error {
	acc, err := handler.getAccountByAddress(params.From)
	if err != nil {
		return err
	}
	err = checkAdminUserSignature(acc, params.Signature, params.PersonalInfo, params.AccessLevel, params.PubKey)
	if err != nil {
		return err
	}
	key, err := utils.ParsePublicKey(params.PubKey)
	if err != nil {
		return err
	}

	accAddress := GetAddressFromPubKey(key)
	_, err = handler.ExecuteTransaction(
		"insert into Accounts (address, personal, level, pkey) values (?, ?, ?, ?)",
		accAddress,
		params.PersonalInfo,
		params.AccessLevel,
		params.PubKey)

	return err
}

// UpdateAccountParams for updating or creating an account
type UpdateAccountParams struct {
	From         Address // who adds the account
	Account      Address // whom to update
	PersonalInfo string  // personal info of the new account
	AccessLevel  int     // access level of the new account
	Signature    []byte  // sender signature
}

// UpdateAccount creates an account
func (handler *AccountHandler) UpdateAccount(params UpdateAccountParams, sucess *bool) error {
	acc, err := handler.getAccountByAddress(params.From)
	err = checkAdminUserSignature(acc, params.Signature, params.Account, params.PersonalInfo, params.AccessLevel)
	if err != nil {
		return err
	}

	handler.ExecuteTransaction(
		"update Accounts set personal=?, level=? where address=?",
		params.PersonalInfo,
		params.AccessLevel,
		params.Account)

	return nil
}

// ListAccounts lists available accounts
func (handler *AccountHandler) ListAccounts() []Account {
	var accounts []Account
	var acc Account
	rows, err := handler.Sp.StateDb.Query("select address, personal, level, pkey from Accounts")
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			pubKeyData := make([]byte, 1024)
			rows.Scan(&acc.Address, &acc.PersonalInfo, &acc.AccessLevel, &pubKeyData)
			acc.PubKey, err = utils.ParsePublicKey(pubKeyData)
			accounts = append(accounts, acc)
		}
	}
	return accounts
}

func (handler *AccountHandler) getAccountByAddress(addr Address) (Account, error) {
	var acc Account
	rows, err := handler.Sp.StateDb.Query("select address, personal, level, pkey from Accounts where address=?", addr)
	defer rows.Close()
	if err != nil {
		return acc, err
	}
	pubKeyData := make([]byte, 1024)
	if rows.Next() {
		rows.Scan(&acc.Address, &acc.PersonalInfo, &acc.AccessLevel, &pubKeyData)
	}
	acc.PubKey, err = utils.ParsePublicKey(pubKeyData)
	return acc, err
}

func checkAdminUserSignature(acc Account, signature []byte, params ...interface{}) error {
	err := acc.PubKey.CheckSignature(
		utils.Hash(params...),
		signature)
	if err != nil {
		return errors.New("invalid user signature")
	}

	if acc.AccessLevel != AdminAccountAccess {
		return errors.New("invalid access level")
	}

	return nil
}

// GetAddressFromPubKey retrieves address from public key
func GetAddressFromPubKey(key utils.SignatureValidator) Address {
	rawData, _ := key.Store()
	return Address(fmt.Sprintf("0x%x", utils.Hash(rawData)[:5]))
}
