package storage

import (
	"log"
)

//StorageProvider handles storing and loading blockchain data from the database
type StorageProvider struct {
	Chain   Blockchain
	ChainDb Database
	StateDb Database
}

//LoadChain loads the chain state from the database.
func (sp *StorageProvider) LoadChain(DbPath string) {
	sp.ChainDb.OpenDb(DbPath + "/blockchain.db")
	sp.StateDb.OpenDb(DbPath + "/storage.db")

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
func (sp *StorageProvider) UpdateChainState() {
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
		err = sp.ChainDb.Transact("INSERT INTO ChainState (id, hash, data) VALUES (?, ?, ?)", item.Id, item.PrevHash, item.Data)
		if err != nil {
			log.Print(err)
		}
	}
}
