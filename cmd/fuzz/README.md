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
