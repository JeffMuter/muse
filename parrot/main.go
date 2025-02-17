package main

import (
	"fmt"
	"log"
	"net/http"
	"parrot/router"
	"time"
)

func main() {
	// setup http server on this port, listens for the aws transcription responses.
	r := router.Router()
	go func() {
		log.Printf("http server is ready")
		log.Fatal(http.ListenAndServe(":8080", r))
	}()

	for {
		time.Sleep(5 * time.Second)
		fmt.Println("hi")
	}
}
