## Usage
### Run Server
#### on a UNIX domain socket
```
go run cmd/fuzz/fuzz.go serve /tmp/fuzz.sock
```
#### over the network on a hostname and port
```
go run cmd/fuzz/fuzz.go serve :8080
go run cmd/fuzz/fuzz.go serve localhost:8080
go run cmd/fuzz/fuzz.go serve 4.2.2.2:8080
```

### Send a single Request
```
go run cmd/fuzz/fuzz.go handshake /tmp/fuzz.sock
go run cmd/fuzz/fuzz.go get_state /tmp/fuzz.sock example_get_state.json
go run cmd/fuzz/fuzz.go handshake localhost:8080
go run cmd/fuzz/fuzz.go get_state localhost:8080 example_get_state.json
```

### Run the test suite
The test suite:
- spawns a Fuzz server and client using an in-memory pipe
- reads the test data under "./test_data"
- sends a sequence of requests using the test_data
- closes both the server, the client, and the underlying connections.

```
go test github.com/New-JAMneration/JAM-Protocol/cmd/fuzz
```

#### With a 5 second timeout
```
go test -timeout 5s github.com/New-JAMneration/JAM-Protocol/cmd/fuzz
```