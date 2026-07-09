# Local Distributed Key-Value Store Simulation
### My first ever distributed System!

## Main Features
1. **Multiple nodes:** This program allows for multiple localhost ports to run concurrently with one another (localhost:8001, localhost:8002, etc.)
2. **Two Phase Commit (2PC):** Includes 2PC for consistency in data in case of failures in backup node using the '/prepare' and '/write' phases
3. **Asynchronous Commit:** Uses go routines to improve efficiency in writes to backup nodes after 2PC
4. **Status Check:** Includes '/status' and '/read' handlers to keep track of memory across all nodes 
5. **Data Persistence:** Writes data to logs after 2PC for persistent storage
6. **Bootstrap Replay** On startup, sees if previous data was stored in logs and writes to database

## How to Run
1. Initialize program and run primary node:
    '''bash
    go run main.go 8001
    '''

2. Initialize and run second, backup node:
    '''bash
    go run main.go 8002
    '''

3. Use "curl" command to write, commit, and test nodes through primary node:
    '''bash
    curl "http://localhost:8001/write?key=user&value=alice"
    '''

## What I learned through this project
As this is my first project using git as well as my first project pertaining to distributed systems, I learned many things:
- The Go language, including (but not limited to) Go's os, fmt, and net/http packages
- Event-driven HTTP Go code, listeners, and Go closures
- Quality and efficient code, freeing up resource leaks using "defer resp.Body.Close()"
- Designing and implementing a distributed system
- What goes into consensus algorithms to handle data flow and prevent split-brain
- How to persistently store data
- How to design bootstrap replay startups