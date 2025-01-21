package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type StatisticsTestCase struct {
	Input     StatisticsInput  `json:"input"`
	PreState  StatisticsState  `json:"pre_state"`
	Output    StatisticsOutput `json:"output"`
	PostState StatisticsState  `json:"post_state"`
}

type StatisticsInput struct {
	Slot        types.TimeSlot       `json:"slot"`
	AuthorIndex types.ValidatorIndex `json:"author_index"`
	Extrinsic   types.Extrinsic      `json:"extrinsic"`
}

type StatisticsOutput struct { // null
}

type StatisticsState struct {
	Pi    types.Statistics     `json:"pi"`
	Tau   types.TimeSlot       `json:"tau"`
	Kappa types.ValidatorsData `json:"kappa_prime"`
}
