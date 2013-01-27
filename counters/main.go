package main

import (
	"database/sql"
	"fmt"
	"time"
)

func main() {
	for t := range time.Tick(time.Second) {
		pg, err := sql.Open("postgres", pgurl)
		if err != nil {
			fmt.Printf("at=error error=%s\n", err)
			return
		}
		if t.Second() == 0 {
			pg.Exec("select reset_minute_counters()")
			fmt.Printf("at=reset bucket=minute time=%s\n", t)
		}
		if t.Minute() == 0 {
			pg.Exec("select reset_hour_counters()")
			fmt.Printf("at=reset bucket=hour time=%s\n", t)
		}
		if t.Hour() == 0 {
			pg.Exec("select reset_day_counters()")
			fmt.Printf("at=reset bucket=day time=%s\n", t)
		}
		pg.Close()
	}
}
