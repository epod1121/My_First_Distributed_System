package main

import (
	"fmt"
	"net/http"
	"os"
)

var database = make(map[string]string)

func main() {
	port := os.Args[1]
	fmt.Println("Starting Program")
	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		writeHandler(w, r, port)
	})
	http.HandleFunc("/prepare", func(w http.ResponseWriter, r *http.Request) {
		prepareHandler(w, r, port)
	})

	fmt.Printf("Starting server on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Server failed to start: ", err)
	}
}

func writeHandler(w http.ResponseWriter, r *http.Request, port string) {
	fmt.Fprintf(w, "This is the write page\n\n")

	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if port == "8001" {
		fmt.Println("Primary: --- Phase 1 ---")

		request := fmt.Sprintf("http://localhost:8002/write?key=%s&value=%s", key, value)
		resp, err := http.Get(request)
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Println("Majority of backups not ready, aborting")
			return
		}
		defer resp.Body.Close()

		fmt.Println("Primary: --- Phase 2 ---")
		prepare := fmt.Sprintf("http://localhost:8002/prepare?key=%s&value=%s", key, value)
		commit := fmt.Sprintf(prepare)
		go http.Get(commit)
	}

	database[key] = value
	fmt.Println("Database: ", database)
}

func prepareHandler(w http.ResponseWriter, r *http.Request, port string) {
	fmt.Printf("Node %s: Received prepare request - readying", port)
	w.WriteHeader(http.StatusOK)
}
