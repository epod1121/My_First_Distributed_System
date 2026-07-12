# Local Distributed Key-Value Store Simulation
### My first ever distributed System!

## Main Features
### Version 1
1. **Multiple nodes:** This program allows for multiple localhost ports to run concurrently with one another (localhost:8001, localhost:8002, etc.)
2. **Two Phase Commit (2PC):** Includes 2PC for consistency in data in case of failures in backup node using the '/prepare' and '/write' phases
3. **Asynchronous Commit:** Uses go routines to improve efficiency in writes to backup nodes after 2PC
4. **Status Check:** Includes '/status' and '/read' handlers to keep track of memory across all nodes
### Version 2
5. **Data Persistence:** Writes data to logs after 2PC for persistent storage
6. **Bootstrap Replay:** On startup, sees if previous data was stored in logs and writes to database
7. **Data Sharding:** Shards data between multiple nodes for efficiency
### Version 3
7. **Time Synchronization:** Uses local time to calculate heartbeats and election timeouts
8. **Leader Election:** Chooses a leader among three nodes (8001, 8002, 8003) that recieves majority of votes



## How to Run
**1. Open three separate terminal windows to launch nodes**
Watch as they find a leader to elect and start sending/recieving hearbeats

```console
go run main.go 8001
```
```console
go run main.go 8002
```
```console
go run main.go 8003
```

**2. Interacting with nodes**

Write data into Shard A (a-m):
```console
curl "http://localhost:8001/write?key=apple&value=red"
```
Write data into Shard B (n-z):
```console
curl "http://localhost:8001/write?key=zebra&value=stripes"
```
Retrieve Shard data
```console
curl "http://localhost:8001/read?key=apple"
```

**Testing a node crash can be simulated buy using Ctrl + C**



## What I learned through this project
As this is my first project using git as well as my first project pertaining to distributed systems, I learned many things:
- The Go language, including Go's os, fmt, net/http, bufio, io, strings, sync, and time packages
- Event-driven HTTP Go code, listeners, and Go closures
- Quality and efficient code, such as freeing up resource leaks using "defer resp.Body.Close()"
- Designing and implementing a distributed system
- What goes into consensus algorithms to handle data flow and prevent split-brain
- How to persistently store data
- How to design bootstrap replay startups
- How to Shard data
- How to use local time in programming distributed systems
- How to implement leader election into a program
