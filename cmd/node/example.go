package main

import (
	"context"

	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/urfave/cli/v3"
)

var exampleCmd = &cli.Command{
	Name: "example",
	Action: func(ctx context.Context, cli *cli.Command) error {
		logger.Info("hello world")
		return nil
	},
}
