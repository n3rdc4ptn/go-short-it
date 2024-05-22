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

// deleteUserCmd represents the deleteUser command
var deleteUserCmd = &cobra.Command{
	Use:   "deleteUser",
	Short: "Deletes a user and its created links from DB",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		db, f, err := internal.GetDB()
		defer f()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		internal.PrepareDB(db)

		err = internal.DeleteUser(db, internal.User{
			Name: args[0],
		})

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Successfully delete user and its data: %v", args[0])
	},
}

func init() {
	rootCmd.AddCommand(deleteUserCmd)
}
