package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// global so it does not have to be passed from main
var (
	database          = make(map[string]string)
	role              = "Follower"
	currentTerm       = 0
	votedFor          = ""
	currentLeader     = ""
	stateMutex        sync.Mutex
	lastHeartbeatTime = time.Now()
)

func main() {
	// reads the port input from the user when initializing node
	port := os.Args[1]
	fmt.Println("Starting Program")

	// creates a file for each port that starts
	filename := fmt.Sprintf("%s.log", port)
	// before all of the handle funcs, open previously logged data (if any)
	// to read and have ready for each node
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	// if the file does not exist, continue on
	if err != nil {
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
	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		readHandler(w, r, port)
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, port)
	})
	http.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		heartbeatHandler(w, r, port)
	})
	http.HandleFunc("/request-vote", func(w http.ResponseWriter, r *http.Request) {
		voteHandler(w, r, port)
	})

	fmt.Printf("Starting server on port %s...\n", port)
	// go routine used to start election timeout counter
	go startElectionTimeout(port)
	// starting node at user given local host port
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Server failed to start: ", err)
	}
}

// used to handle key/value init from url
// 2 phase commit
// as well as handle primary node (in this case node 8001
// was hard coded in for simplicity when testing)
func writeHandler(w http.ResponseWriter, r *http.Request, port string, filename string) {

	// stores key
	key := r.URL.Query().Get("key")

	// set a port to target all traffic to
	targetPort := "8002"

	// if the user enters a key, get the first letter
	if len(key) > 0 {
		firstLetter := key[0]

		// if the first letter falls in the second half of the alphabet
		if firstLetter >= 'n' && firstLetter <= 'z' {
			// reroute that traffic to the 3rd node (sharding!)
			targetPort = "8003"
		}
	}

	// stores value
	value := r.URL.Query().Get("value")

	// node 8001 is primary node
	if port == "8001" {
		fmt.Println("Primary: --- Phase 1 ---")

		// reaches out to target port to write key and value
		// as well as sees if it is ready to commit to storage
		prepare := fmt.Sprintf("http://localhost:%s/prepare?key=%s&value=%s", targetPort, key, value)
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
		commit := fmt.Sprintf("http://localhost:%s/write?key=%s&value=%s", targetPort, key, value)
		// writes the key value pair to storage
		// go routine used so primary node does not have to wait for backup's writing to storage
		go http.Get(commit)
		fmt.Println("Primary: --- 2PC Complete ---\n")
	}

	// opens the file associated with the port and edits it based off of flags
	// O_CREATE creates the file if it does not exist
	// O_WRONLY writes to the file if it is empty
	// O_APPEND writes at the end of the file if there is data already written
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	// checks for error when opening
	if err != nil {
		fmt.Println("Error opening file")
	}

	if port != "8001" {
		// adds to database
		database[key] = value

		// closes file to prevent corruption and leaking
		defer file.Close()
		fmt.Fprintf(file, "%s,%s\n", key, value)
		// prints database to show consistency
		fmt.Println("Database: ", database)
	}
}

// used to make sure a node is ready to commit a write
func prepareHandler(w http.ResponseWriter, r *http.Request, port string) {
	fmt.Printf("Node %s: Received prepare request - readying\n", port)
	// sends back 200 code
	w.WriteHeader(http.StatusOK)
}

// used to make sure the key exists
func readHandler(w http.ResponseWriter, r *http.Request, port string) {
	// gets the key from the URL
	key := r.URL.Query().Get("key")

	// first statement handles leader fetching
	// if the port is the leader
	if port == "8001" {

		// sets a target port
		targetPort := "8002"

		// if the key starts with a letter in the second half of the alphabet
		// change the target port to the port that holds that hashed key
		if len(key) > 0 && key[0] >= 'n' && key[0] <= 'z' {
			targetPort = "8003"
		}

		// stores the target port URL
		shardURL := fmt.Sprintf("http://localhost:%s/read?key=%s", targetPort, key)

		// sends the reques to get it, if it is not available print so
		resp, err := http.Get(shardURL)
		if err != nil {
			http.Error(w, "Shard unreachable", http.StatusInternalServerError)
			return
		}
		// close to prevent leaking
		defer resp.Body.Close()

		// writes the value
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}

	// if the value is not in the
	value, exists := database[key]
	if !exists {
		http.Error(w, "Key not found: ", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Value: %s\n", value)
}

// shows database contents so it is easy to see if something is off
func statusHandler(w http.ResponseWriter, r *http.Request, port string) {
	fmt.Fprintf(w, "Port: %s\nDatabase contents: %v\n", port, database)
}

// handles heartbeats
func heartbeatHandler(w http.ResponseWriter, r *http.Request, port string) {
	// gets the leader value from the URL
	leaderPort := r.URL.Query().Get("leader")

	// later lock and unlock statement as a placeholder
	stateMutex.Lock()
	defer stateMutex.Unlock()

	// prints confirmation of heatbeat recieved
	fmt.Printf("Node %s: Recieved heartbeat from leader %s\n", port, leaderPort)

	// role change
	currentLeader = leaderPort
	role = "Follower"

	// adds global heartbeat time
	lastHeartbeatTime = time.Now()
	// confirmation of working
	w.WriteHeader(http.StatusOK)
}

// handles votes
func voteHandler(w http.ResponseWriter, r *http.Request, port string) {
	candidate := r.URL.Query().Get("candidate")

	stateMutex.Lock()
	defer stateMutex.Unlock()

	fmt.Printf("Node %s received vote request from Candidate %s\n", port, candidate)

	if role == "Follower" {
		w.WriteHeader(http.StatusOK)
		fmt.Printf("Node %s voted YES for %s\n", port, candidate)
		return
	}

	w.WriteHeader(http.StatusNotModified)
}

func startElectionTimeout(port string) {
	for {
		time.Sleep(200 * time.Millisecond)

		stateMutex.Lock()
		if role == "Leader" {
			stateMutex.Unlock()
			continue
		}

		timeSinceLastHeartbeat := time.Since(lastHeartbeatTime)

		if timeSinceLastHeartbeat > 3 * time.Second {
			fmt.Printf("Node %s: Master has timed out - No heartbeat for %v.\nStarting election\n", port, timeSinceLastHeartbeat)
			launchElection(port)
		}

		stateMutex.Unlock()
	}
}

func launchElection(port string) {
	role = "Candidate"
	currentTerm++
	votedFor = port
	votesReceived := 1
	fmt.Printf("Node %s: Starting term %d - voting for self\n", port, currentTerm)

	allPorts := []string{"8001 ", "8002", "8003"}

	for _, peerPort := range allPorts {
		if peerPort == port {
			continue
		}

		voteURL := fmt.Sprintf("http:/localhost:%s/request-vote?candidate=%s&term=%d", peerPort, port, currentTerm)

		resp, err := http.Get(voteURL)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			votesReceived++
		}
	}

	if votesReceived >= 2 {
		fmt.Printf("Node %s won the election with %d votes\n", port, votesReceived)
		role = "Leader"
		currentLeader = port
		go startHeartbeatTicker(port)
	} else {
		fmt.Printf("Node %s failed election - lost majority with %d votes", port, votesReceived)
		role = "Follower"
	}
}

func startHeartbeatTicker(port string){
	ticker := time.NewTicker(1 * time.Second)
	allPorts := []string{"8001", "8002", "8003"}

	for range ticker.C {
		stateMutex.Lock()
		if role != "Leader" {
			stateMutex.Unlock()
			ticker.Stop()
			return
		}
		stateMutex.Unlock()

		for _, peerPort := range allPorts {
			if peerPort == port {
				continue
			}

			heartbeatURL := fmt.Sprintf("http://localhost:%s/heartbeat?leader=%s", peerPort, port)
			go http.Get(heartbeatURL)
		}
	}
}