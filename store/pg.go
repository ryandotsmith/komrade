package store

import (
	"database/sql"
	"fmt"
	"github.com/bmizerany/pq"
	"os"
)

var (
	pg *sql.DB
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
		fmt.Printf("error=%s\n", err)
		os.Exit(1)
	}
}
