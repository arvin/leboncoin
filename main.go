//go:generate go run generate_embed.go
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = `dbname=leboncoin sslmode=disable`
	}

	flag.Parse()
	switch flag.Arg(0) {
	case "fetch":
		fetch(dbURL)

	default:
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		log.Fatal(http.ListenAndServe(":"+port, server(dbURL)))
	}
}

type Announce struct {
	Href, Title, Department, City, District string
	Price                                   *int
	PublishedAt                             time.Time `db:"published_at"`
	Urgent                                  bool
}
