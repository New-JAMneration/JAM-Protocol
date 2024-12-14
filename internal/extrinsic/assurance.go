package extrinsic

import "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"

// AvailAssurance defined in jam_types
type AvailAssurance jam_types.AvailAssurance

// SetCoreBit is for setting the validity of the core-index in the assurance bit
func (a AvailAssurance) SetCoreBit(coreIndex uint16, validity bool) {
	// equation 11.15 ~ 11.16
	// 先判定 core 有 pending available 的 report 和 timeout (timeout 在 jam_types line 101 已經處理)
	// 由 workreport get core-index，並驗證 validity
}

// GenerateMessage is for generating the message, which is along with the assurance signature
func GenerateMessage() {
	// equation 11.13,  currently lack of serialization & encoding & signing context

}
