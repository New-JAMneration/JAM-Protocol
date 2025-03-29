package PolkaVM

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

type RefineInput struct {
	WorkItemIndex       uint               // i
	WorkPackage         types.WorkPackage  // p
	AuthOutput          types.ByteSequence // o
	ImportSegments      []types.ImportSpec // bold{i}
	ExportSegmentOffset uint               // zeta
}

type RefineOutput struct {
	WorkResult    types.WorkExecResult
	WorkReport    types.WorkReport
	ImportSegment types.ImportSpec
	Gas           types.Gas
}

// B.4 M
type IntegratedPVMType struct {
	ProgramCode ProgramCode    // p
	Memory      Memory         // u
	PC          ProgramCounter // i
}

func RefineInvoke(input RefineInput) RefineOutput {
	return RefineOutput{}
}
