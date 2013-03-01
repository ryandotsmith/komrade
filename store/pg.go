package store

import (
	"database/sql"
	"fmt"
	"github.com/bmizerany/pq"
	"log"
	"os"
)

var (
	pg      *sql.DB
	statsPg *sql.DB
)

func parseUrl(name string) string {
	var tmp string
	var err error

	tmp = os.Getenv(name)
	if len(tmp) == 0 {
		fmt.Printf("at=error error=\"must set %s\"\n", name)
		os.Exit(1)
	}

	tmp, err = pq.ParseURL(tmp)
	if err != nil {
		fmt.Printf("at=error error=\"unable to parse %s\"\n", name)
		os.Exit(1)
	}
	return tmp
}

func init() {
	var err error

	pg, err = sql.Open("postgres", parseUrl("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	statsPg, err = sql.Open("postgres", parseUrl("STATS_DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
}
