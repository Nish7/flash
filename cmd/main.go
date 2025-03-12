package main

import (
	server "github.com/nish7/flash/internal"
	"log"
)

func main() {
	server := server.NewServer(":8085", server.NewInMemoryStore())
	err := server.Start()
	if err != nil {
		log.Fatalf("Error: Starting the server %v", err)
	}
}
