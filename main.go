package main

import (
	"github.com/New-JAMneration/JAM-Protocol/cmd"
	"github.com/New-JAMneration/JAM-Protocol/config"
)

func main() {
	config.InitConfig()

	// 啟動程式
	cmd.Execute()
}
