## Usage
### Run Server
```
go run cmd/fuzz/fuzz.go serve /tmp/fuzz.sock
```

### Send the Request
```
go run cmd/fuzz/fuzz.go handshake /tmp/fuzz.sock
go run cmd/fuzz/fuzz.go get_state /tmp/fuzz.sock example_get_state.json
go run cmd/fuzz/fuzz.go test_folder "$SOCKET_PATH" "$FOLDER_PATH"
```