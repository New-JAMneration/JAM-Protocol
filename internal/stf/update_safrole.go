package stf

import (
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
)

func UpdateSafrole() error {
	start := time.Now()

	err := safrole.OuterUsedSafrole()
	if err != nil {
		return err
	}

	end := time.Now()

	duration := end.Sub(start)
	fmt.Printf("\033[31mSafrole update took %s\033[0m\n", duration)

	return nil
}
