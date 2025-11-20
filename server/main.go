package main

import (
	"log"
	"net/http"
	"os"
)

type mineRequest struct {
	Data string `json:"data"`
}

func main() {
	port := envOrDefault("PORT", "8080")
	node := newNode()

	log.Printf("blockchain node listening on :%s", port)
	if err := http.ListenAndServe(":"+port, node.handler()); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
