## Usage

```
go run ./cmd/fuzz --help
```

### Run jam-test-vectors

```
go run ./cmd/node test --help
```

### Run Server
```
go run ./cmd/fuzz /tmp/fuzz.sock
```

### Send the Request
```
go run ./cmd/fuzz handshake /tmp/fuzz.sock
go run ./cmd/fuzz get_state /tmp/fuzz.sock example_get_state.json
go run ./cmd/fuzz test_folder "$SOCKET_PATH" "$FOLDER_PATH"
```