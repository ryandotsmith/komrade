package main

import (
	"database/sql"
	"fmt"
	"time"
)

func main() {
	go resetMinute()
	select {}
}

func resetMinute() {
	for t := range time.Tick(time.Minute) {
		pg, err := sql.Open("postgres", pgurl)
		if err != nil {
			fmt.Printf("at=error error=%s\n", err)
			return
		}
		pg.Exec("select reset_minute_counters()")
		fmt.Printf("at=reset bucket=minute time=%s\n", t)
		pg.Close()
	}
}
