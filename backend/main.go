package main

import (
	"github.com/pogzyb/api"
	"log"
)

func main() {
	addr := ":8080"
	log.Printf("Starting API @ %s\n", addr)
	api.Run(addr)
}
