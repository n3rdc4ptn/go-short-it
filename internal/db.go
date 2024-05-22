package internal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/tursodatabase/go-libsql"
)

func GetDB() (*sql.DB, func(), error) {
	dbName := "local.db"
	primaryUrl := viper.GetString("turso_database_url")
	authToken := viper.GetString("turso_auth_token")
	syncInterval := time.Minute

	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		fmt.Println("Error creating temporary directory:", err)
		return nil, nil, err
	}

	dbPath := filepath.Join(dir, dbName)

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl, libsql.WithAuthToken(authToken), libsql.WithSyncInterval(syncInterval))
	if err != nil {
		fmt.Println("Error creating connector:", err)
		return nil, nil, err
	}

	db := sql.OpenDB(connector)

	return db, func() {
		os.RemoveAll(dir)
		connector.Close()
		db.Close()
	}, nil
}

func PrepareDB(db *sql.DB) {
	db.Exec(`CREATE TABLE IF NOT EXISTS links (
		slug TEXT PRIMARY KEY,
		target TEXT NOT NULL,
		created_by TEXT NOT NULL
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
		name TEXT PRIMARY KEY
	)`)
}

func DropDB(db *sql.DB) {
	db.Exec(`DROP TABLE links`)
	db.Exec(`DROP TABLE users`)
}

func CreateUser(db *sql.DB, user User) (err error) {
	_, err = db.Exec(`INSERT INTO users (name) VALUES (?) ON CONFLICT(name) DO NOTHING`, user.Name)
	return
}

func DeleteUser(db *sql.DB, user User) (err error) {
	_, err = db.Exec(`DELETE FROM users WHERE name = ?`, user.Name)
	if err != nil {
		return
	}

	_, err = db.Exec(`DELETE FROM links WHERE created_by = ?`, user.Name)
	return
}

func GetUser(db *sql.DB, name string) (user User, err error) {
	row := db.QueryRow(`SELECT name FROM users WHERE name = ?`, name)

	if err = row.Scan(&user.Name); err != nil {
		return
	}

	return
}

type User struct {
	Name string
}

type Link struct {
	Slug      string `uri:"slug" form:"slug" json:"slug"`
	Target    string `form:"target" json:"target"`
	CreatedBy string
}
