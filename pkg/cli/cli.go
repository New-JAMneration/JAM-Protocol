package cli

import (
	"fmt"
	"os"
)

var Sub = make(map[string]*Command)

func Run() {
	args := os.Args[1:]

	if len(args) == 0 {
		ShowShortHelp()
		return
	}

	if args[0] == "help" || args[0] == "--help" {
		ShowLongHelp()
		return
	}

	if cmd, ok := Sub[args[0]]; ok {
		strings := args[1:]
		cmd.preRunCmd(strings)
	} else {
		fmt.Printf("Unknown command: %s\n\n", args[0])
		ShowShortHelp()
		os.Exit(1)
	}
}

func ShowShortHelp() {
	fmt.Printf("Usage: [command]\n\n")
	fmt.Println("Commands:")

	maxLen := 0
	for name, _ := range Sub {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	for name, cmd := range Sub {
		fmt.Printf("  %-*s  %s\n", maxLen, name, cmd.Short)
	}
}

func ShowLongHelp() {
	fmt.Printf("Usage: [command]\n\n")
	fmt.Println("Commands:")

	maxLen := 0
	for name, _ := range Sub {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	for name, cmd := range Sub {
		fmt.Printf("  %-*s  %s\n", maxLen, name, cmd.Long)
	}
}
