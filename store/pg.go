package store

import (
	"fmt"
	"github.com/bmizerany/pq"
	"os"
)

var (
	pgurl string
)

func init() {
	var err error
	url := os.Getenv("DATABASE_URL")
	if len(url) == 0 {
		fmt.Printf("at=error error=\"must set DATABASE_URL\"\n")
		os.Exit(1)
	}
	pgurl, err = pq.ParseURL(url)
	if err != nil {
		fmt.Printf("at=error error=\"unable to parse DATABASE_URL\"\n")
		os.Exit(1)
	}
}
