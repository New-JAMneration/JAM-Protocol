package cli

import (
	"flag"
	"fmt"
	"log"
)

type Command struct {
	Use   string
	Short string
	Long  string
	Run   func(args []string)
	Flags []Flag
}

func (c *Command) preRunCmd(args []string) {
	var help bool

	c.parseFlags(args, &help)

	if help {
		c.ShowHelp()
		return
	}

	c.Run(args)
}

func (c *Command) ShowHelp() {
	fmt.Printf("Usage: %s \n", c.Use)
	fmt.Println(c.Long)
	fmt.Println("Flags:")

	maxLen := 0
	for _, f := range c.Flags {
		if len(f.NameStr()) > maxLen {
			maxLen = len(f.NameStr())
		}
	}

	for _, f := range c.Flags {
		fmt.Printf("  --%-*s  %s\n", maxLen, f.NameStr(), f.String())

	}
}

func (c *Command) parseFlags(args []string, help *bool) {
	if len(args) == 0 || len(c.Flags) == 0 {
		return
	}

	set := flag.NewFlagSet(c.Use, flag.ExitOnError)
	set.BoolVar(help, "help", false, "show help")

	for _, f := range c.Flags {
		f.Apply(set)
	}

	err := set.Parse(args)
	if err != nil {
		log.Fatal(err)
	}
}

func AddCommand(cmd *Command) {
	Sub[cmd.Use] = cmd
}
