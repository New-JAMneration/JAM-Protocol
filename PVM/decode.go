package PVM

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

func getRegModIndex(instructionCode []byte, pc ProgramCounter) uint8 {
	return min(12, (instructionCode[pc+1])%16)
}

func getRegFloorIndex(instructionCode []byte, pc ProgramCounter) uint8 {
	return min(12, (instructionCode[pc+1])>>4)
}

// A.5.2
func decodeOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, error) {
	lX := min(4, skipLength)
	immediateData := instructionCode[pc+1 : pc+lX+1]
	immediate, _, err := ReadUintSignExtended(immediateData, len(immediateData))
	if err != nil {
		return 0, err
	}
	return int(immediate), nil
}

// A.5.3
func decodeOneRegisterAndOneExtendedWidthImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, uint64, error) {
	panic("not implemented")
}

// A.5.4
func decodeTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint64, uint64, error) {
	lX := ProgramCounter(min(4, uint8(instructionCode[pc+1])))

	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pc+2+lX], types.U64(lX))
	if err != nil {
		return 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d deserialize vx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	vX, err := SignExtend(uint8(lX), uint64(decodedVX))
	if err != nil {
		return 0, 0, fmt.Errorf("opcosde %s(%d) at pc=%d signExtend lx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	lY := min(4, max(0, skipLength-lX-1))
	decodedVy, err := utils.DeserializeFixedLength(instructionCode[pc+2+lX:pc+2+lX+lY], types.U64(lY))
	if err != nil {
		return 0, 0, fmt.Errorf("opcosde %s(%d) at pc=%d deserialization vy raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}
	vY, err := SignExtend(uint8(lY), uint64(decodedVy))
	if err != nil {
		return 0, 0, fmt.Errorf("opcosde %s(%d) at pc=%d signExtend lx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	return vX, vY, nil
}

// A.5.5
// returns vX
func decodeOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (ProgramCounter, error) {
	lX := min(4, skipLength)
	offsetData := instructionCode[pc+1 : pc+1+lX]
	offset, _, err := ReadIntFixed(offsetData, len(offsetData))
	if err != nil {
		return 0, err
	}

	return pc + ProgramCounter(offset), nil
}

// A.5.6
// returns rA, vX
func decodeOneRegisterAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint64, error) {
	rA := min(12, instructionCode[pc+1]%16)
	lX := min(4, max(0, skipLength-1))

	immediateData := instructionCode[pc+2 : pc+2+lX]
	immediate, _, err := ReadUintSignExtended(immediateData, len(immediateData))
	if err != nil {
		pvmLogger.Errorf("opcode %s at instruction %d deserialize vy raise error : %s", zeta[opcode(instructionCode[pc])], pc, err)
		return 0, 0, err
	}

	return rA, immediate, nil
}

// A.5.7
func decodeOneRegisterAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int8, uint64, uint64, error) {
	rA := int8(min(12, instructionCode[pc+1]%16))
	lX := min(4, ProgramCounter(uint8((instructionCode[pc+1] >> 4))))
	pcMargin := pc + 2 + lX
	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pcMargin], types.U64(lX))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d deserialize vx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}
	vX, err := SignExtend(uint8(lX), uint64(decodedVX))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d signExtend vx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	lY := min(4, max(0, skipLength-lX-1))
	decodedVY, err := utils.DeserializeFixedLength(instructionCode[pcMargin:pcMargin+lY], types.U64(lY))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d deserialize vy raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}
	vY, err := SignExtend(uint8(lY), uint64(decodedVY))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d signExtend vy raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	return rA, vX, vY, nil
}

// A.5.8
// returns rA, vX, vY
func decodeOneRegisterOneImmediateAndOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint64, ProgramCounter, error) {
	rA := min(12, instructionCode[pc+1]%16)
	lX := ProgramCounter(min(4, (instructionCode[pc+1]>>4)%8))
	lY := min(4, max(0, skipLength-lX-1))

	immediateData := instructionCode[pc+2 : pc+2+lX]
	immediate, _, err := ReadUintSignExtended(immediateData, len(immediateData))
	if err != nil {
		return 0, 0, 0, err
	}

	offsetData := instructionCode[pc+2+lX : pc+2+lX+lY]
	offset, _, err := ReadIntFixed(offsetData, len(offsetData))
	if err != nil {
		return 0, 0, 0, err
	}

	return rA, immediate, pc + ProgramCounter(offset), nil
}

// A.5.9
func decodeTwoRegisters(instructionCode []byte, pc ProgramCounter) (rD uint8, rA uint8, err error) {
	if int(pc+1) >= len(instructionCode) {
		return 0, 0, errors.New("pc out of bound")
	}
	rD = getRegModIndex(instructionCode, pc)
	rA = getRegFloorIndex(instructionCode, pc)
	return rD, rA, nil
}

func decodeTwoRegistersAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint8, uint64, error) {
	rA := min(12, instructionCode[pc+1]&15)
	rB := min(12, instructionCode[pc+1]>>4)
	lX := min(4, max(0, skipLength-1))
	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pc+2+lX], types.U64(lX))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d deserialization error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}
	vX, err := SignExtend(uint8(lX), uint64(decodedVX))
	if err != nil {
		return 0, 0, 0, err
	}

	return rA, rB, vX, nil
}

// A.5.11
// returns rA, rB, vX
func decodeTwoRegistersAndOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint8, ProgramCounter, error) {
	rA := min(12, instructionCode[pc+1]%16)
	rB := min(12, instructionCode[pc+1]>>4)
	lX := min(4, max(0, skipLength-1))

	offsetData := instructionCode[pc+2 : pc+2+lX]
	offset, _, err := ReadIntFixed(offsetData, len(offsetData))
	if err != nil {
		return 0, 0, 0, err
	}

	return rA, rB, pc + ProgramCounter(offset), nil
}

// A.5.12
// returns rA, rB, vX, vY
func decodeTwoRegistersAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint8, uint64, uint64, error) {
	rA := min(12, instructionCode[pc+1]%16)
	rB := min(12, instructionCode[pc+1]>>4)
	lX := ProgramCounter(min(4, instructionCode[pc+2]%8))
	lY := min(4, max(0, skipLength-lX-2))

	vXData := instructionCode[pc+3 : pc+3+lX]
	vX, _, err := ReadUintFixed(vXData, len(vXData))
	if err != nil {
		return 0, 0, 0, 0, err
	}

	vYData := instructionCode[pc+3+lX : pc+3+lX+lY]
	vY, _, err := ReadUintFixed(vYData, len(vYData))
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return rA, rB, vX, vY, nil
}

// A.5.13
func decodeThreeRegisters(instructionCode []byte, pc ProgramCounter) (rA uint8, rB uint8, rD uint8, err error) {
	if int(pc+2) >= len(instructionCode) {
		return 0, 0, 0, errors.New("pc out of bound")
	}
	rA = getRegModIndex(instructionCode, pc)
	rB = getRegFloorIndex(instructionCode, pc)
	rD = min(12, instructionCode[pc+2])
	return rA, rB, rD, nil
}

func storeIntoMemory(interp *Interpreter, offset int, memIndex uint32, immediate uint64) ExitReason {
	mem := interp.Memory
	if memIndex < uint32(1<<16) { // 0.7.2  A.8 check memory > 2^16
		return ExitPanic
	}

	pageNum := memIndex / ZP
	pageIndex := memIndex % ZP

	page, ok := mem.Pages[pageNum]
	if !ok {
		return ExitPageFault | ExitReason(memIndex)
	}
	if page.Access != MemoryReadWrite {
		return ExitPageFault | ExitReason(memIndex)
	}

	// Serialize little-endian directly to a stack buffer — no heap alloc.
	var buf [8]byte
	switch offset {
	case 1:
		buf[0] = byte(immediate)
	case 2:
		binary.LittleEndian.PutUint16(buf[:2], uint16(immediate))
	case 4:
		binary.LittleEndian.PutUint32(buf[:4], uint32(immediate))
	case 8:
		binary.LittleEndian.PutUint64(buf[:8], immediate)
	}
	src := buf[:offset]

	// Fast path: entirely within current page.
	if pageIndex+uint32(offset) <= ZP {
		copy(page.Value[pageIndex:], src)
		return ExitContinue
	}

	// Cross-page slow path.
	nextPage, ok := mem.Pages[pageNum+1]
	if !ok {
		return ExitPageFault | ExitReason(memIndex)
	}
	if nextPage.Access != MemoryReadWrite {
		return ExitPageFault | ExitReason(memIndex)
	}

	firstLen := ZP - pageIndex
	copy(page.Value[pageIndex:], src[:firstLen])
	copy(nextPage.Value, src[firstLen:])
	return ExitContinue
}

/*
func storeIntoMemory(interp *Interpreter, offset int, memIndex uint32, Immediate uint64) ExitReason {
	mem := interp.Memory
	if memIndex < uint32(1<<16) { // 0.7.2  A.8 check memory > 2^16
		return ExitPanic
	}
	vX := uint32(memIndex)
	pageNum := vX / ZP
	pageIndex := memIndex % ZP

	vY := utils.SerializeFixedLength(types.U64(Immediate), types.U64(offset))

	if _, pageAllocated := mem.Pages[pageNum]; pageAllocated {
		// try to allocate read-only memory --> page-fault
		if mem.Pages[pageNum].Access != MemoryReadWrite {
			return ExitPageFault | ExitReason(memIndex)
		}
		if pageIndex+uint32(len(vY)) <= ZP { // data allocated do not exceed maxSize
			copy(mem.Pages[pageNum].Value[pageIndex:], vY)
		} else { // data allocated exceed maxSize --> cross page
			// check next page access
			if _, pageAllocated := mem.Pages[pageNum+1]; !pageAllocated {
				return ExitPageFault | ExitReason(memIndex)
			}

			currentPageLength := ZP - pageIndex
			currentPageData := vY[:currentPageLength]
			nextPageData := vY[currentPageLength:]

			copy(mem.Pages[pageNum].Value[pageIndex:], currentPageData)
			copy(mem.Pages[pageNum+1].Value[:len(nextPageData)], nextPageData)
		}

	} else { // page not allocated
		return ExitPageFault | ExitReason(memIndex)
	}
	return ExitContinue
}
*/

/*
	func loadFromMemory(interp *Interpreter, offset uint32, vx uint32) (uint64, ExitReason) {
		mem := interp.Memory
		if vx < uint32(1<<16) { // 0.7.2  A.8 check memory > 2^16
			return 0, ExitPanic
		}
		vX := uint32(vx)

		pageNum := vX / ZP
		pageIndex := vX % ZP
		// load memory : the page must be exist and at least readable
		// we allocated memory at least read-only
		// => if the memory is not allocated -> it's Inaccessible
		if _, pageAllocated := mem.Pages[pageNum]; !pageAllocated {
			return 0, ExitPageFault | ExitReason(vX)
		}
		memBytes := make([]byte, offset)

		if pageIndex+offset <= ZP {
			memBytes = mem.Pages[pageNum].Value[pageIndex : pageIndex+offset]
		} else { // cross page memory memory loading
			if mem.Pages[pageNum+1] == nil {
				return 0, ExitPageFault | ExitReason(vX)
			}

			remainBytes := mem.Pages[pageNum+1].Value[:offset-(ZP-pageIndex)]
			copy(memBytes, mem.Pages[pageNum].Value[pageIndex:]) // copy current page
			copy(memBytes[ZP-pageIndex:], remainBytes)           // copy next page
		}
		memVal, err := utils.DeserializeFixedLength(memBytes, types.U64(offset))
		if err != nil {
			pvmLogger.Errorf("loadFromMemory deserialization error %s", err)
			return 0, ExitPanic
		}
		return uint64(memVal), ExitContinue
	}
*/
func loadFromMemory(interp *Interpreter, offset uint32, vx uint32) (uint64, ExitReason) {
	mem := interp.Memory
	if vx < uint32(1<<16) { // 0.7.2  A.8 check memory > 2^16
		return 0, ExitPanic
	}

	pageNum := vx / ZP
	pageIndex := vx % ZP

	page, ok := mem.Pages[pageNum]
	if !ok {
		return 0, ExitPageFault | ExitReason(vx)
	}

	// Fast path: load fully within a single page — no alloc, no per-byte loop.
	if pageIndex+offset <= ZP {
		v := page.Value[pageIndex : pageIndex+offset]
		switch offset {
		case 1:
			return uint64(v[0]), ExitContinue
		case 2:
			return uint64(binary.LittleEndian.Uint16(v)), ExitContinue
		case 4:
			return uint64(binary.LittleEndian.Uint32(v)), ExitContinue
		case 8:
			return binary.LittleEndian.Uint64(v), ExitContinue
		}
	}

	// Cross-page slow path: assemble bytes into a stack buffer.
	nextPage, ok := mem.Pages[pageNum+1]
	if !ok {
		return 0, ExitPageFault | ExitReason(vx)
	}

	var buf [8]byte
	firstLen := ZP - pageIndex
	copy(buf[:firstLen], page.Value[pageIndex:])
	copy(buf[firstLen:offset], nextPage.Value[:offset-firstLen])

	switch offset {
	case 2:
		return uint64(binary.LittleEndian.Uint16(buf[:2])), ExitContinue
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf[:4])), ExitContinue
	case 8:
		return binary.LittleEndian.Uint64(buf[:8]), ExitContinue
	}
	return 0, ExitPanic // unreachable: offset ∈ {1,2,4,8}
}
