package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// global so it does not have to be passed from main
var database = make(map[string]string)

func main() {
	// reads the port input from the user when initializing node
	port := os.Args[1]
	fmt.Println("Starting Program")

	// creates a file for each port that starts
	filename := fmt.Sprintf("%s.log", port)
	// before all of the handle funcs, open previously logged data (if any)
	// to read and have ready for each node
	file, exist := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	// if the file does not exist, continue on
	if exist != nil {
		fmt.Println("No file found")
		// if the file does exist, read the contents of it
	} else {
		// open scanner
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// get data line by line from the file
			line := scanner.Text()

			// parses up the data from the files while scanning
			parts := strings.Split(line, ",")
			// writes the data to the database
			database[parts[0]] = parts[1]
		}
	}

	// handle funcs for different node tools and passing port to the handlers

	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		writeHandler(w, r, port, filename)
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
func writeHandler(w http.ResponseWriter, r *http.Request, port string, filename string) {

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
		commit := fmt.Sprintf("http://localhost:8002/write?key=%s&value=%s", key, value)
		// writes the key value pair to storage
		// go routine used so primary node does not have to wait for backup's writing to storage
		go http.Get(commit)
	}

	// adds to database
	database[key] = value

	// opens the file associated with the port and edits it based off of flags
	// O_CREATE creates the file if it does not exist
	// O_WRONLY writes to the file if it is empty
	// O_APPEND writes at the end of the file if there is data already written
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	// checks for error when opening
	if err != nil {
		fmt.Println("Error opening file")
	}
	// closes file to prevent corruption and leaking
	defer file.Close()

	fmt.Fprintf(file, "%s,%s\n", key, value)

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
