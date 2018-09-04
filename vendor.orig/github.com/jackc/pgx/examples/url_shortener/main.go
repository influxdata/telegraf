package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/log15adapter"
	log "gopkg.in/inconshreveable/log15.v2"
)

var pool *pgx.ConnPool

// afterConnect creates the prepared statements that this application uses
func afterConnect(conn *pgx.Conn) (err error) {
	_, err = conn.Prepare("getUrl", `
    select url from shortened_urls where id=$1
  `)
	if err != nil {
		return
	}

	_, err = conn.Prepare("deleteUrl", `
    delete from shortened_urls where id=$1
  `)
	if err != nil {
		return
	}

	_, err = conn.Prepare("putUrl", `
    insert into shortened_urls(id, url) values ($1, $2)
    on conflict (id) do update set url=excluded.url
  `)
	return
}

func getUrlHandler(w http.ResponseWriter, req *http.Request) {
	var url string
	err := pool.QueryRow("getUrl", req.URL.Path).Scan(&url)
	switch err {
	case nil:
		http.Redirect(w, req, url, http.StatusSeeOther)
	case pgx.ErrNoRows:
		http.NotFound(w, req)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func putUrlHandler(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path
	var url string
	if body, err := ioutil.ReadAll(req.Body); err == nil {
		url = string(body)
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err := pool.Exec("putUrl", id, url); err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func deleteUrlHandler(w http.ResponseWriter, req *http.Request) {
	if _, err := pool.Exec("deleteUrl", req.URL.Path); err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func urlHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		getUrlHandler(w, req)

	case "PUT":
		putUrlHandler(w, req)

	case "DELETE":
		deleteUrlHandler(w, req)

	default:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	logger := log15adapter.NewLogger(log.New("module", "pgx"))

	var err error
	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     "127.0.0.1",
			User:     "jack",
			Password: "jack",
			Database: "url_shortener",
			Logger:   logger,
		},
		MaxConnections: 5,
		AfterConnect:   afterConnect,
	}
	pool, err = pgx.NewConnPool(connPoolConfig)
	if err != nil {
		log.Crit("Unable to create connection pool", "error", err)
		os.Exit(1)
	}

	http.HandleFunc("/", urlHandler)

	log.Info("Starting URL shortener on localhost:8080")
	err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Crit("Unable to start web server", "error", err)
		os.Exit(1)
	}
}
