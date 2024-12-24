package cmd

import (
    "github.com/New-JAMneration/JAM-Protocol/pkg/cli"
    "github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func Execute() {
    // Initialize global store
    _ = store.GetInstance()
    
    // Run the CLI
    cli.Run()
}