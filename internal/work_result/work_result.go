package work_result

type WorkExecResultType string

const (
	WorkExecResultOk           WorkExecResultType = "ok"
	WorkExecResultOutOfGas                        = "out-of-gas"
	WorkExecResultPanic                           = "panic"
	WorkExecResultBadCode                         = "bad-code"
	WorkExecResultCodeOversize                    = "code-oversize"
)

type WorkExecResult map[WorkExecResultType][]byte

func GetWorkExecResult(resultType WorkExecResultType, data []byte) WorkExecResult {
	if resultType == WorkExecResultOk {
		return map[WorkExecResultType][]byte{
			resultType: data,
		}
	}

	return map[WorkExecResultType][]byte{
		resultType: nil,
	}
}

// WorkResult (14.4)
type WorkResult struct {
	ServiceId     uint32         `json:"service_id,omitempty"`
	CodeHash      [32]byte       `json:"code_hash,omitempty"`
	PayloadHash   [32]byte       `json:"payload_hash,omitempty"`
	AccumulateGas uint64         `json:"accumulate_gas,omitempty"`
	Result        WorkExecResult `json:"result,omitempty"`
}
