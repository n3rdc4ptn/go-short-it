/*
Copyright © 2024 Moritz Reich
*/
package cmd

import (
	"fmt"
	"os"

	"dev.moritzreich.shortit/internal"
	"github.com/spf13/cobra"
)

var resetDBCmd = &cobra.Command{
	Use:   "resetDB",
	Short: "A brief description of your command",

	Run: func(cmd *cobra.Command, args []string) {
		db, f, err := internal.GetDB()
		defer f()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		internal.DropDB(db)
		internal.PrepareDB(db)
	},
}

func init() {
	rootCmd.AddCommand(resetDBCmd)
}
