package main

import (
	"flag"
	"github.com/New-JAMneration/JAM-Protocol/cmd"
	"github.com/New-JAMneration/JAM-Protocol/config"
	"os"
)

func main() {
	initConfig()

	// 啟動程式
	cmd.Execute()
}

func initConfig() {
	var configPath string
	set := flag.NewFlagSet("config", flag.ExitOnError)
	set.StringVar(&configPath, "config", "example.json", "set config file path")
	err := set.Parse(os.Args[1:])
	if err != nil {
		return
	}

	if err := config.LoadConfig(configPath); err != nil {
		panic(err)
	}
}
