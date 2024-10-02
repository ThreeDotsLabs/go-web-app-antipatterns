package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	for {
		_, err := net.DialTimeout("tcp", "postgres:5432", time.Second)
		if err == nil {
			return
		}
		fmt.Println("postgres not up yet, retrying...")
		time.Sleep(time.Second * 5)
	}
}
