package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type Mentira struct {
	ID   int64
	Code string
	URL  string
}

func main() {
	rand.Seed(time.Now().UnixNano())

	connStr := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable",
		"mentira",
		"mentira",
		"mentira")
	openDbConnection(connStr)

	http.HandleFunc("/_/create", createHandler)
	http.HandleFunc("/", redirectHandler)
	panic(http.ListenAndServe(":8000", nil))
}

var db *sql.DB

func openDbConnection(connStr string) {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	code := generateCode(6)

	_, err := db.Exec("INSERT INTO mentira (code, url) VALUES ($1, $2)", code, url)
	if err != nil {
		panic(err)
	}
}

var validCodeChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func generateCode(size int) string {
	b := make([]rune, size)
	for i := range b {
		b[i] = validCodeChars[rand.Intn(len(validCodeChars))]
	}
	return string(b)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]

	var mentira Mentira
	row := db.QueryRow("SELECT * FROM mentira WHERE code = $1", code)
	err := row.Scan(&mentira.ID, &mentira.Code, &mentira.URL)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, mentira.URL, http.StatusFound)
}
