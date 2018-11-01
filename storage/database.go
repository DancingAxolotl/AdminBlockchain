package storage

import (
	"database/sql"
	"errors"
	"log"

	//blank import
	_ "github.com/mattn/go-sqlite3"
)

// Database convinient wrapper for accessing sql DB
type Database struct {
	database *sql.DB
}

// OpenDb opens a specified database
func (db *Database) OpenDb(path string) {
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	db.database = database
}

// Close closes the database connection
func (db *Database) Close() {
	db.database.Close()
}

// IsOpen returns true if the database connection has been established
func (db *Database) IsOpen() bool {
	return db.database != nil
}

// Transact performs a transaction on the database
func (db *Database) Transact(statement string, params ...interface{}) error {
	if !db.IsOpen() {
		return errors.New("database not loaded")
	}

	_, err := db.database.Exec(statement, params...)

	return err
}

// Query performs a query on the database
func (db *Database) Query(query string) (*sql.Rows, error) {
	if db.database == nil {
		return nil, errors.New("database not loaded")
	}
	rows, err := db.database.Query(query)
	return rows, err
}
