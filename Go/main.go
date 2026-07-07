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
	database[key] = value

	if port == "8001" {
		fmt.Println("Primary node: Sending request")

		request := fmt.Sprintf("http://localhost:8002/write?key=%s&value=%s", key, value)
		resp, err := http.Get(request)
		if err != nil {
			fmt.Println("Failed to send request: ", err)
		}
		defer resp.Body.Close()
	}

	fmt.Println("Database: ", database)
}
