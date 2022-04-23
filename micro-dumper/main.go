package main

import (
	"fmt"
	"os"
	"redislog"
	"time"
)

func main() {
	fmt.Println("start to retrieve request record ...")
	f, _ := os.Create("./request.log")
	for {
		time.Sleep(1)
		if record, found := redislog.ReadRecord(); found {
			_, err := f.WriteString(fmt.Sprintf("remote addr: %s url: %s access time: %d time executed: %d body bytes sent: %d\n", record.RemoteAddr, record.URL, record.AccessTime, record.TimeExecuted, record.BodyBytesSent))
			if err != nil {
				panic(err)
			} else {
				fmt.Println("write a request record to log file")
			}
		}
	}
}
