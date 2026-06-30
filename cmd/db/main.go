package cmd

import (
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Short: "A Perfectly Normal Wheel Server with associated database management commands",
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return db, nil
}
