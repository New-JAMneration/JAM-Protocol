package cmd

import (
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/pkg/cli"
)

func init() {
	var exampleCmd = &cli.Command{
		Use:   "example",
		Short: "Prints 'hello world",
		Long:  "Prints 'hello world",
		Run: func(args []string) {
			fmt.Println("hello world")
		},
	}


	cli.AddCommand(exampleCmd)
}
