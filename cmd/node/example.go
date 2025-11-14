package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var exampleCmd = &cli.Command{
	Name: "example",
	Action: func(ctx context.Context, cli *cli.Command) error {
		fmt.Println("hello world")
		return nil
	},
}
