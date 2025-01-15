package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type AuthorizationTestCase struct {
	Input     AuthorizationInput  `json:"input"`
	PreState  AuthorizationState  `json:"pre_state"`
	Output    AuthorizationOutput `json:"output"`
	PostState AuthorizationState  `json:"post_state"`
}

/*
This sequence is out of GP spec and derived from the Guarantees Extrinsic (E_G)

For the sake of construction simplicity, we prefer not to include the complete
	extrinsic here but rather focus only on the components relevant to advancing this subsystem's STF.

	CoreAuthorizers[i] = CoreAuthorizer {
    	core: E_G[i].w.core,
    	auth-hash: E_G[i].w.auth-hash
	}
*/
type CoreAuthorizer struct {
	CoreIndex      types.CoreIndex  `json:"core"`
	AuthorizerHash types.OpaqueHash `json:"auth_hash"`
}

type CoreAuthorizers []CoreAuthorizer

type AuthorizationInput struct {
	Slot  types.TimeSlot  `json:"slot"`
	Auths CoreAuthorizers `json:"auths"`
}

type AuthorizationOutput struct { // null
}

type AuthorizationState struct {
	Alpha  types.AuthPools  `json:"auth_pools"`
	Varphi types.AuthQueues `json:"auth_queues"`
}
