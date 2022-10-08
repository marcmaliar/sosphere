// Package main defines the backend webserver, with routes,
// handlers, templates, helpers, and middleware.
package main

import (
	"database/sql"  // create database pool, sql go interface
	"flag"          // command line arguments
	"html/template" // template cache
	"log"           // custom loggers
	"net/http"      // http Server
	"os"            // os streams
	"time"

	"person.mmaliar.com/internal/models" // Database model

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4" // formDecoder
	_ "github.com/go-sql-driver/mysql" // mysql driver
)

// application holds important objects that we want to inject into methods
// using dependency injection
type application struct {
	infoLog        *log.Logger
	errorLog       *log.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

// main parses flags, makes database pool, creates application struct,
// initializes router, and runs http Server
func main() {

	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:password@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.IdleTimeout = 12 * time.Hour
	sessionManager.Store = mysqlstore.New(db)

	app := &application{
		infoLog:        infoLog,
		errorLog:       errorLog,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s...", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// openDB opens a mysql Database Connection Pool,
// returning an error if the database pool could not be opened
// or it's not available for connection.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil { // Check if database is alright
		return nil, err
	}
	return db, nil
}
