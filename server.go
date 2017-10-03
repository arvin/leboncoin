package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func server(dbURL string) http.Handler {
	db := sqlx.MustConnect("postgres", dbURL)
	t, err := loadTemplates()
	if err != nil {
		log.Fatal(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var dest []Announce
		if err := db.Select(&dest, queries["select.sql"]); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := t.ExecuteTemplate(w, "index.html", dest); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func loadTemplates() (*template.Template, error) {
	t := template.New("")
	for k, v := range templates {
		_, err := t.New(k).Parse(v)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
