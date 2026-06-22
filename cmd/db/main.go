package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/ncruces/go-sqlite3/driver"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return db, nil
}
