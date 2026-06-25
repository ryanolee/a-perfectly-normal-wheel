package cmd

import (
	"fmt"

	internaldb "github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
	"github.com/spf13/cobra"
)

var seedDBPath string

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Clear the database and fill it with seed data",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB(seedDBPath)
		if err != nil {
			return err
		}
		defer db.Close()

		if _, err := db.Exec(internaldb.SeedSQL); err != nil {
			return fmt.Errorf("apply seed: %w", err)
		}

		cmd.Printf("database seeded at %s\n", seedDBPath)
		return nil
	},
}

func init() {
	seedCmd.Flags().StringVar(&seedDBPath, "db", "data.db", "path to the SQLite database file")
	RootCmd.AddCommand(seedCmd)
}
