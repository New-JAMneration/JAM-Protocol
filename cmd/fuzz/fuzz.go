package main

import (
	"context"
	"log"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
)

func printUsage() {
	usage := `Usage: go run cmd/fuzz/fuzz.go COMMAND [ARGS...]

Valid commands are:
  serve FILE
  handshake FILE
  import_block FILE - not yet implemented
  set_state FILE - not yet implemented
  get_state FILE - not yet implemented
  help [COMMAND]`

	log.Fatalln(usage)
}

type Handler func([]string)

var handlers = map[string]Handler{
	"serve":        serve,
	"handshake":    handshake,
	"import_block": importBlock,
	"set_state":    setState,
	"get_state":    getState,
	"help":         help,
}

func main() {
	if len(os.Args) == 1 {
		printUsage()
	}

	handler, valid := handlers[os.Args[1]]
	if !valid {
		printUsage()
	}

	config.InitConfig()

	handler(os.Args[2:])
}

func serve(args []string) {
	if len(args) == 0 {
		helpImpl("serve")
	}

	server, err := fuzz.NewFuzzServer("unix", args[0])
	if err != nil {
		log.Fatalf("error creating server: %v\n", err)
	}

	server.ListenAndServe(context.Background())
}

func handshake(args []string) {
	if len(args) == 0 {
		helpImpl("handshake")
	}

	client, err := fuzz.NewFuzzClient("unix", args[0])
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}

	defer client.Close()

	var info fuzz.PeerInfo

	if err := info.FromConfig(); err != nil {
		log.Fatalf("error reading config: %v\n", err)
	}

	resp, err := client.Handshake(info)
	if err != nil {
		log.Fatalf("error sending request: %v\n", err)
	}

	log.Println("received handshake response:")
	log.Printf("  Name: %s\n", resp.Name)
	log.Printf("  App Version: %v\n", resp.AppVersion)
	log.Printf("  JAM Version: %v\n", resp.JamVersion)
}

func importBlock(args []string) {
	// TODO
}

func setState(args []string) {
	// TODO
}

func getState(args []string) {
	// TODO
}

func help(args []string) {
	helpImpl(args...)
}

func helpImpl(args ...string) {
	if len(args) == 0 {
		printUsage()
	}

	switch args[0] {
	case "serve":
		log.Fatalln("serve FILE - starts a server listening on FILE via named Unix socket")
	case "handshake":
		log.Fatalln("handshake FILE - connects to a server listening on FILE and sends a handshake")
	case "import_block":
		log.Fatalln("import_block FILE - connects to a server listening on FILE and sends an import_block request")
	case "set_state":
		log.Fatalln("set_state FILE - connects to a server listening on FILE and sends a set_state request")
	case "get_state":
		log.Fatalln("get_state FILE - connects to a server listening on FILE and sends a get_state request")
	default:
		printUsage()
	}
}
