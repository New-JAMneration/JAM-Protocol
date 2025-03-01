package PolkaVM

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// 定義 General Function 的回傳類型
type GeneralFunctionOutput struct {
	ExitReason error                     // exit reason
	GasRemain  Gas                       // gas remain
	Register   Registers                 // new registers
	Ram        Memory                    // new memory
	State      types.ServiceAccountState // new state
	Addition   any                       // addition host-call context
}

type GeneralFunctionInput struct {
	Gas       Gas
	Registers Registers
	Memory    Memory
	State     types.ServiceAccountState
	Args      []any // For different type F()
}

// General Function type
type GeneralFunction func(GeneralFunctionInput) GeneralFunctionOutput
