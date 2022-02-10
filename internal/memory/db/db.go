package db

import (
	"database/sql"
	"errors"
	"linkShortener/internal/memory"
	"linkShortener/pkg/helpers"

	_ "github.com/lib/pq"
)

type database struct {
	memory.Memory
	source *sql.DB
}

var dbSingleInstance database

func GetDBInstance() database {
	if dbSingleInstance.source == nil {
		db, err := sql.Open("postgres", helpers.GetConfig().DbConnString)
		if err != nil {
			panic("Database open error: " + err.Error())
		}

		err = db.Ping()
		if err != nil {
			panic("Database ping error: " + err.Error())
		}

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS links (
			id SERIAL PRIMARY KEY,
		   longLink TEXT NOT NULL,
		   author TEXT NOT NULL,
		   createdAt timestamp)`)
		if err != nil {
			panic("Database creating table links error: " + err.Error())
		}
		dbSingleInstance = database{
			source: db,
		}
	}
	return dbSingleInstance
}

func (d database) AddEntry(entry memory.MemoryRequest) (int64, error) {

	err := d.source.QueryRow(`INSERT INTO links (longLink, author, createdAt) VALUES ($1, $2, timestamp $3) returning id;`, entry.Long, entry.Author, entry.CreatedAt.Format("2006-01-02 15:04:05")).Scan(&entry.Short)
	if err != nil {
		return 0, err
	}
	return entry.Short, nil
}
func (d database) GetEntry(entry memory.MemoryRequest) (memory.MemoryRequest, error) {
	var row *sql.Row
	if entry.Short == 0 {
		row = d.source.QueryRow(`select * from links where longLink = $1;`, entry.Long)
	} else {
		row = d.source.QueryRow(`select * from links where id = $1;`, entry.Short)
	}
	if row.Err() != nil {
		return entry, row.Err()
	}
	response := memory.MemoryRequest{}
	err := row.Scan(&response.Short, &response.Long, &response.Author, &response.CreatedAt)
	if err != nil && err.Error() == "sql: no rows in result set" {
		return response, errors.New("No such entry")
	}
	if err != nil {
		return response, err
	}
	return response, nil
}
func (d database) Clear() error {
	return nil
}
