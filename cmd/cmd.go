package cmd

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/pkg/cli"
)

func Execute() {
	// Initialize global store
	_ = store.GetInstance()

	// Run the CLI
	cli.Run()
}
