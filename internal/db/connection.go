package db

import (
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
)

const (
	dbFilePath = "data.db"
)

type DBConfig struct {
	FilePath string
}

func NewDBConnection(config DBConfig) (*Queries, error) {
	sqlDB, err := sql.Open("sqlite3", config.FilePath)
	if err != nil {
		return nil, err
	}

	return New(sqlDB), nil
}
