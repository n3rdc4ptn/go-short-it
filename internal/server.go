package internal

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var Release string

func RunServer() {
	if Release == "" {
		fmt.Println("Dev build")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	db, f, err := GetDB()
	defer f()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	PrepareDB(db)

	r := SetupRouter(db)

	port := viper.GetInt("port")

	r.Run("0.0.0.0:" + fmt.Sprint(port)) // listen and serve on 0.0.0.0:8080
}
