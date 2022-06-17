package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	dbHost := getenv("DB_HOST", "localhost")
	dbPort := getenv("DB_PORT", "5432")
	dbUser := getenv("DB_USER", "mentira")
	dbPass := getenv("DB_PASS", "mentira")
	dbName := getenv("DB_NAME", "mentira")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	openDbConnection(connStr)

	http.HandleFunc("/_/health", healthHandler)
	http.HandleFunc("/_/create", createHandler)
	http.HandleFunc("/", redirectHandler)
	panic(http.ListenAndServe(":8000", nil))
}

func getenv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
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

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		db.Close()
		fmt.Println("Closed connection, Walison!!")
		os.Exit(0)
	}()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
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
