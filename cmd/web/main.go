package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"html/template"
	"log"
	"net/http"
	"os"

	"deathnote.owner.lalamilight/internal/models"

	"github.com/go-playground/form/v4"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	notes          *models.NoteModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")

	//get the database connection URL
	dbUrl := "postgresql://postgres:Zhak159*@localhost:5432/pgx_death_note"

	flag.Parse()
	// 2 loggers
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	dbPool, err := pgxpool.Connect(context.Background(), dbUrl)
	if err != nil {
		fmt.Println("failed to connect to postgresql", err)
		return
	}
	// to close DB pool
	defer dbPool.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	//Decoder instance
	formDecoder := form.NewDecoder()

	// scs.New() function initialize a new session manager.
	//Then we configure it to use our database as the session store
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(dbPool)
	sessionManager.Lifetime = 12 * time.Hour

	// init a new application struct
	// Initialize a models.NoteModel instance and add it to the application dependencies.
	//add templateCache
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		notes:          &models.NoteModel{DB: dbPool},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Init a new http.Server struct
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	// Using new 2 loggers
	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe() // call new http.Server struct
	errorLog.Fatal(err)
}
