/*
Copyright © 2024 Moritz Reich
*/
package cmd

import (
	"fmt"
	"os"

	"dev.moritzreich.shortit/internal"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createTokenCmd represents the createToken command
var createTokenCmd = &cobra.Command{
	Use:   "createToken",
	Short: "Create a token for a user",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		db, f, err := internal.GetDB()
		defer f()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		internal.PrepareDB(db)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user": args[0],
		})

		secret := viper.GetString("app_secret")
		tokenString, err := token.SignedString([]byte(secret))

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Token: %v", tokenString)
	},
}

func init() {
	rootCmd.AddCommand(createTokenCmd)
}
