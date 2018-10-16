package storage

import (
	"log"
)

//StorageProvider handles storing and loading blockchain data from the database
type StorageProvider struct {
	chain   Blockchain
	stateDb Database
}

//LoadChain loads the chain state from the database.
func (sp StorageProvider) LoadChain(stateDbPath string) {
	sp.stateDb.OpenDb(stateDbPath)

	rows, err := sp.stateDb.Query("SELECT * FROM ChainState")
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
		sp.chain.InsertBlock(Block{id, hash, data})
	}
	rows.Close()

	if !sp.chain.IsValid() {
		log.Fatal("The chain state database is corrupt.")
	}
}

//UpdateChainState writes the current blockchain into the database
func (sp StorageProvider) UpdateChainState() {
	if !sp.stateDb.IsOpen() {
		log.Fatal("The chain state database is not available")
	}

	sp.stateDb.Transact("CREATE TABLE IF NOT EXISTS ChainState (id integer, hash blob, data text)")

	rows, err := sp.stateDb.Query("SELECT COUNT(*) FROM ChainState")
	if err != nil {
		log.Fatal(err)
	}
	var count int
	rows.Scan(&count)

	for _, item := range sp.chain[count:] {
		err = sp.stateDb.Transact("INSERT INTO ChainState (id, hash, data) VALUES (?, ?, ?)", item.id, item.prevHash, item.data)
		if err != nil {
			log.Print(err)
		}
	}
}
