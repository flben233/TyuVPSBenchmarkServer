package repo

import (
	"database/sql"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func InitDatabase() {
	db, err := sql.Open("sqlite3", "index.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS report (
            id VARCHAR PRIMARY KEY, 
            path VARCHAR,
            date VARCHAR
        );
    `)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS speedtest (
            spot VARCHAR,
            download FLOAT,
            upload FLOAT,
            rid VARCHAR,
            PRIMARY KEY (spot, rid)
        );
    `)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS route (
            spot VARCHAR,
            rtype VARCHAR,
            rid VARCHAR,
            PRIMARY KEY (spot, rid)
        );
    `)
	if err != nil {
		panic(err)
	}
}

func GetDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "index.db")
	if err != nil {
		panic(err)
	}
	return db
}
