package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tursodatabase/go-libsql"
)

var Release string

func main() {
	if Release == "" {
		fmt.Println("Dev build")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	dbName := "local.db"
	primaryUrl := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	syncInterval := time.Minute

	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		fmt.Println("Error creating temporary directory:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, dbName)

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl, libsql.WithAuthToken(authToken), libsql.WithSyncInterval(syncInterval))
	if err != nil {
		fmt.Println("Error creating connector:", err)
		os.Exit(1)
	}
	defer connector.Close()

	db := sql.OpenDB(connector)
	defer db.Close()

	// db.Exec("DROP TABLE links")
	db.Exec(`CREATE TABLE IF NOT EXISTS links (
		slug TEXT PRIMARY KEY,
		target TEXT NOT NULL
	)`)

	r := gin.Default()

	// Redirect to target
	r.GET("/:slug", func(c *gin.Context) {
		var link Link

		if err := c.ShouldBindUri(&link); err != nil {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}

		row := db.QueryRow("SELECT * FROM links WHERE slug = ?", link.Slug)
		if err := row.Scan(&link.Slug, &link.Target); err != nil {
			c.JSON(404, gin.H{"message": "Not found"})
			return
		}

		c.Redirect(307, link.Target)
	})

	r.GET("/", func(c *gin.Context) {
		rows, err := db.Query("SELECT * FROM links")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to execute query: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		var links []Link = make([]Link, 0)

		for rows.Next() {
			var link Link

			if err := rows.Scan(&link.Slug, &link.Target); err != nil {
				fmt.Println("Error scanning row:", err)
				return
			}

			links = append(links, link)
		}

		if err := rows.Err(); err != nil {
			fmt.Println("Error during rows iteration:", err)
		}

		c.JSON(200, links)
	})

	r.POST("/", func(c *gin.Context) {
		var link Link
		c.Bind(&link)

		fmt.Printf("%v", link)
		db.Exec(`INSERT INTO links (slug, target) VALUES (?,?)
			ON CONFLICT(slug) DO UPDATE SET target=?
		`, link.Slug, link.Target, link.Target)

		c.JSON(200, link)
	})

	r.DELETE("/:slug", func(c *gin.Context) {
		var link Link

		if err := c.ShouldBindUri(&link); err != nil {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}

		_, err := db.Exec("DELETE FROM links WHERE slug = ?", link.Slug)
		if err != nil {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "Deleted " + link.Slug,
		})

	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

type Link struct {
	Slug   string `uri:"slug" form:"slug" json:"slug"`
	Target string `form:"target" json:"target"`
}
