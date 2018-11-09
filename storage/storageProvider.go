package storage

import (
	"log"
)

var (
	// StateDbPath path to the state database
	StateDbPath string
)

//Provider handles storing and loading blockchain data from the database
type Provider struct {
	Chain   Blockchain
	ChainDb Database
	StateDb Database
}

//LoadChain loads the chain state from the database.
func (sp *Provider) LoadChain(DbPath string) {
	sp.ChainDb.OpenDb(DbPath + "/blockchain.db")
	const stateDbName = "/storage.db"
	StateDbPath = DbPath + stateDbName

	sp.ChainDb.Transact("CREATE TABLE IF NOT EXISTS ChainState (id integer, hash blob, data text)")

	rows, err := sp.ChainDb.Query("SELECT * FROM ChainState")
	if err != nil {
		log.Fatal(err)
	}
	var id int
	var hash []byte
	var data string

	for rows.Next() {
		err = rows.Scan(&id, &hash, &data) //integer, blob, text
		if err != nil {
			log.Fatal(err)
		}
		sp.Chain.InsertBlock(Block{id, hash, data})
	}
	rows.Close()

	if !sp.Chain.IsValid() {
		log.Fatal("The chain state database is corrupted.")
	}
}

//UpdateChainState writes the current blockchain into the database
func (sp *Provider) UpdateChainState() {
	rows, err := sp.ChainDb.Query("SELECT MAX(id) FROM ChainState")
	if err != nil {
		log.Fatal(err)
	}

	var count int
	if rows.Next() {
		rows.Scan(&count)
	}
	rows.Close()

	for _, item := range sp.Chain[count:] {
		err = sp.ChainDb.Transact("INSERT INTO ChainState (id, hash, data) VALUES (?, ?, ?)", item.ID, item.PrevHash, item.Data)
		if err != nil {
			log.Print(err)
		}
	}
}
