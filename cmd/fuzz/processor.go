package main

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
)

type TestProcessor interface {
	ScanFolder(folderPath string) ([]string, error)
	ProcessFile(client *fuzz.FuzzClient, filePath string) error
	ProcessInitialize(client *fuzz.FuzzClient) error
	ProcessImportBlock(client *fuzz.FuzzClient) error
}

type OptionalPeerInfoProcessor interface {
	ProcessPeerInfo(client *fuzz.FuzzClient) error
}
