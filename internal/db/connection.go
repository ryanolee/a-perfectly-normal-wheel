package db

import (
	"database/sql"
	"log"

	_ "github.com/ncruces/go-sqlite3/driver"
)

const (
	dbFilePath = "data.db"
)

type DBConfig struct {
	FilePath string
}

func NewDBConnection(config DBConfig) *Queries {
	sqlDB, err := sql.Open("sqlite3", config.FilePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	return New(sqlDB)
}
