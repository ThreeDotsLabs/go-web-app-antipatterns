package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	for {
		_, err := net.DialTimeout("tcp", "mysql:3306", time.Second)
		if err == nil {
			return
		}
		fmt.Println("mysql not up yet, retrying...")
		time.Sleep(time.Second * 5)
	}
}
