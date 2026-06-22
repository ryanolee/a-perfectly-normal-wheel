package main

import (
	"fmt"

	internaldb "github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
	"github.com/spf13/cobra"
)

var migrateDBPath string

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply the database schema to a SQLite file",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB(migrateDBPath)
		if err != nil {
			return err
		}
		defer db.Close()

		if _, err := db.Exec(internaldb.SchemaSQL); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}

		cmd.Printf("schema applied to %s\n", migrateDBPath)
		return nil
	},
}

func init() {
	migrateCmd.Flags().StringVar(&migrateDBPath, "db", "data.db", "path to the SQLite database file")
	rootCmd.AddCommand(migrateCmd)
}
