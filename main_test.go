package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func prepareLocalDB() (*sql.DB, func(), error) {
	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		return nil, func() {
			os.RemoveAll(dir)
		}, err
	}

	db, err := sql.Open("libsql", "file:"+dir+"/test.db")
	if err != nil {
		return nil, func() {
			os.RemoveAll(dir)
		}, err
	}

	return db, func() {
		os.RemoveAll(dir)
	}, nil
}

func TestPrepareDB(t *testing.T) {
	db, closing_f, err := prepareLocalDB()
	defer closing_f()
	defer func() {
		if closeError := db.Close(); closeError != nil {
			fmt.Println("Error closing database", closeError)
			if err == nil {
				t.Error(err)
			}
		}
	}()

	if err != nil {
		t.Error(err)
	}

	prepare_db(db)

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		t.Error(err)
	}

	var tables []struct {
		name string
	}
	for rows.Next() {
		var table struct {
			name string
		}

		if err := rows.Scan(&table.name); err != nil {
			fmt.Println("Error scanning row:", err)
		}

		tables = append(tables, table)
	}

	if len(tables) != 1 {
		// It should only be one table
		t.Error("There is more or less then one table")
	}

	if tables[0].name != "links" {
		t.Error("The created table is not called links")
	}
}

func TestListLinkRoute(t *testing.T) {
	db, closing_f, err := prepareLocalDB()
	defer closing_f()
	defer func() {
		if closeError := db.Close(); closeError != nil {
			fmt.Println("Error closing database", closeError)
			if err == nil {
				t.Error(err)
			}
		}
	}()

	if err != nil {
		t.Error(err)
	}

	prepare_db(db)

	gin.SetMode(gin.TestMode)
	router := setupRouter(db)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "[]", w.Body.String())

	w = httptest.NewRecorder()

	exampleLink := Link{
		Slug:   "test",
		Target: "testtarget",
	}
	linkJson, _ := json.Marshal(exampleLink)
	req, _ = http.NewRequest("POST", "/", strings.NewReader(string(linkJson)))
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(linkJson), w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	exampleList := make([]Link, 0)
	exampleList = append(exampleList, exampleLink)

	listJson, _ := json.Marshal(exampleList)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(listJson), w.Body.String())
}

func TestCreateLinkRoute(t *testing.T) {
	db, closing_f, err := prepareLocalDB()
	defer closing_f()
	defer func() {
		if closeError := db.Close(); closeError != nil {
			fmt.Println("Error closing database", closeError)
			if err == nil {
				t.Error(err)
			}
		}
	}()

	if err != nil {
		t.Error(err)
	}

	prepare_db(db)

	gin.SetMode(gin.TestMode)
	router := setupRouter(db)

	w := httptest.NewRecorder()

	exampleLink := Link{
		Slug:   "test",
		Target: "testtarget",
	}
	linkJson, _ := json.Marshal(exampleLink)
	req, _ := http.NewRequest("POST", "/", strings.NewReader(string(linkJson)))
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(linkJson), w.Body.String())

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

	assert.Equal(t, len(links), 1)

	assert.Equal(t, links[0], exampleLink)

	w = httptest.NewRecorder()

	req, _ = http.NewRequest("GET", "/"+exampleLink.Slug, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(linkJson), w.Body.String())
}

func TestDeleteLinkRoute(t *testing.T) {
	db, closing_f, err := prepareLocalDB()
	defer closing_f()
	defer func() {
		if closeError := db.Close(); closeError != nil {
			fmt.Println("Error closing database", closeError)
			if err == nil {
				t.Error(err)
			}
		}
	}()

	if err != nil {
		t.Error(err)
	}

	prepare_db(db)

	gin.SetMode(gin.TestMode)
	router := setupRouter(db)

	w := httptest.NewRecorder()

	exampleLink := Link{
		Slug:   "test",
		Target: "testtarget",
	}

	db.Exec(`INSERT INTO links (slug, target) VALUES (?,?)
			ON CONFLICT(slug) DO UPDATE SET target=?
	`, exampleLink.Slug, exampleLink.Target, exampleLink.Target)

	req, _ := http.NewRequest("DELETE", "/"+exampleLink.Slug, nil)
	router.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}
