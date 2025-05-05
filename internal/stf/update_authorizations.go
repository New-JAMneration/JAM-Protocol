package stf

import "github.com/New-JAMneration/JAM-Protocol/internal/authorization"

func UpdateAuthorizations() error {
	// === Run Authorization ===
	err := authorization.Authorization()
	if err != nil {
		return err
	}
	return nil
}
