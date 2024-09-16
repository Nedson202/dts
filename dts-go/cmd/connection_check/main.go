package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <host> <port>")
		os.Exit(1)
	}

	host := os.Args[1]
	port := os.Args[2]

	for i := 0; i < 30; i++ { // Try for 30 seconds
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Second)
		if err == nil {
			conn.Close()
			fmt.Printf("Successfully connected to %s:%s\n", host, port)
			os.Exit(0)
		}
		time.Sleep(time.Second)
	}

	fmt.Printf("Failed to connect to %s:%s after 30 attempts\n", host, port)
	os.Exit(1)
}
