package cmd

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/sum"
	"github.com/New-JAMneration/JAM-Protocol/pkg/cli"
)

func init() {
	var a, b int

	var sumCmd = &cli.Command{
		Use:   "sum",
		Short: "Calculate the sum of two numbers",
		Long:  "The 'sum' command calculates the sum of two numbers provided using -a and -b flags.",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:         "a",
				Usage:        "First number",
				DefaultValue: 0,
				Destination:  &a,
			},
			&cli.IntFlag{
				Name:         "b",
				Usage:        "Second number",
				DefaultValue: 0,
				Destination:  &b,
			},
		},
		Run: func(args []string) {
			sum.Sum(a, b)
		},
	}

	cli.AddCommand(sumCmd)
}
