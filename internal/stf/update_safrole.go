package stf

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
)

func UpdateSafrole() error {
	err := safrole.OuterUsedSafrole()
	if err != nil {
		return err
	}
	return nil
}
