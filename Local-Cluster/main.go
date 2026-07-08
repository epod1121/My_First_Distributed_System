package main

import (
	"fmt"
	"net/http"
	"os"
)

// global so it does not have to be passed from main
var database = make(map[string]string)

func main() {
	// reads the port input from the user when initializing node
	port := os.Args[1]
	fmt.Println("Starting Program")

	// handle funcs for different node tools and passing port to the handlers

	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		writeHandler(w, r, port)
	})
	http.HandleFunc("/prepare", func(w http.ResponseWriter, r *http.Request) {
		prepareHandler(w, r, port)
	})
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, port)
	})

	fmt.Printf("Starting server on port %s...\n", port)
	// starting node at user given local host port
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Server failed to start: ", err)
	}
}

// used to handle key/value init from url
// 2 phase commit
// as well as handle primary node (in this case node 8001 was hard coded in for simplicity when testing)
func writeHandler(w http.ResponseWriter, r *http.Request, port string) {

	// gets key and value stored
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	// node 8001 is primary node
	if port == "8001" {
		fmt.Println("Primary: --- Phase 1 ---")

		// reaches out to node 8002 to write key and value
		// as well as sees if it is ready to commit to storage
		prepare := fmt.Sprintf("http://localhost:8002/prepare?key=%s&value=%s", key, value)
		resp, err := http.Get(prepare)
		// checks if backup is ready
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Println("Backups not ready, aborting")
			http.Error(w, "Aborted: ", http.StatusInternalServerError)
			return
		}
		// closed so no resource leak
		defer resp.Body.Close()

		// phase 2 of 2PC
		fmt.Println("Primary: --- Phase 2 ---")
		// writes the key value pair to storage
		commit := fmt.Sprintf("http://localhost:8002/write?key=%s&value=%s", key, value)
		// go routine used so primary node does not have to wait for backup's writing to storage
		go http.Get(commit)
	}

	database[key] = value
	// prints database to show consistency
	fmt.Println("Database: ", database)
}

// used to make sure a node is ready to commit a write
func prepareHandler(w http.ResponseWriter, r *http.Request, port string) {
	fmt.Printf("Node %s: Received prepare request - readying\n", port)
	// sends back 200 code
	w.WriteHeader(http.StatusOK)
}

// used to make sure the key exists
func readHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value, exists := database[key]

	if !exists {
		http.Error(w, "Key not found: ", http.StatusNotFound)
	} else {
		fmt.Fprintf(w, "Key found: %v\n", value)
	}
}

// shows database contents so it is easy to see if something is off
func statusHandler(w http.ResponseWriter, r *http.Request, port string) {
	fmt.Fprintf(w, "Port: %s\nDatabase contents: %v\n", port, database)
}
