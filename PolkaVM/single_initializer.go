package PolkaVM

import "fmt"

type StandardCodeFormat []byte // p
type Argument []byte           // a
type Instructions []byte       // c

func P(x int) uint32 {
	return ZP * ((uint32(x) + ZP - 1) / ZP)
}

func Z(x int) uint32 {
	return ZZ * ((uint32(x) + ZZ - 1) / ZZ)
}

func SingleInitializer(p StandardCodeFormat, a Argument) (Instructions, Registers, Memory, error) {

	c, o, w, z, s, err := DecodeSerializedValues(p)
	if err != nil {
		return nil, Registers{}, Memory{}, err
	}
	if 5*ZZ+uint64(Z(len(o)))+uint64(Z(len(w)+int(z)*int(ZP)+int(s)+ZI)) > 1<<32 {
		return nil, Registers{}, Memory{}, fmt.Errorf("Memory layout calculations failed")
	}

	// Memory layout calculations
	readOnlyStart := uint32(ZZ)
	readOnlyEnd := readOnlyStart + uint32(len(o))
	readOnlyPadding := readOnlyStart + P(len(o))
	readWriteStart := 2*ZZ + Z(len(o))
	readWriteEnd := readWriteStart + uint32(len(w))
	readWritePadding := readWriteStart + P(len(w)) + uint32(z)*ZP
	stackEnd := uint32(1<<32 - 2*ZZ - ZI)
	stackStart := stackEnd - P(int(s))
	argumentStart := stackEnd
	argumentEnd := argumentStart + uint32(len(a))
	argumentPadding := argumentStart + P(len(a))

	mem := Memory{Pages: make(map[uint32]*Page)}

	allocateMemorySegment(&mem, readOnlyStart, readOnlyEnd, o, MemoryReadOnly)
	allocateMemorySegment(&mem, readOnlyEnd, readOnlyPadding, nil, MemoryReadOnly) // Padding

	allocateMemorySegment(&mem, readWriteStart, readWriteEnd, w, MemoryReadWrite)
	allocateMemorySegment(&mem, readWriteEnd, readWritePadding, nil, MemoryReadWrite) // Padding

	allocateMemorySegment(&mem, argumentStart, argumentEnd, a, MemoryReadOnly)
	allocateMemorySegment(&mem, argumentEnd, argumentPadding, nil, MemoryReadOnly) // Padding

	allocateStack(&mem, stackStart, stackEnd)

	// Registers initialization
	var regs Registers
	regs[0] = uint64(1<<32 - 1<<16)
	regs[1] = uint64(1<<32 - 2*ZZ - ZI)
	regs[7] = uint64(1<<32 - ZZ - ZI)
	regs[8] = uint64(len(a))

	return c, regs, mem, nil
}

func allocateMemorySegment(mem *Memory, start, end uint32, content []byte, access MemoryAccess) {
	for addr := start; addr < end; addr += ZP {
		pageNum := addr / ZP
		// check page overlap for padding
		if content == nil {
			if _, exists := mem.Pages[pageNum]; exists {
				continue
			} else {
				mem.Pages[pageNum] = &Page{
					Value:  make([]byte, ZP),
					Access: access,
				}
				continue
			}
		}
		pageSize := min(len(content), int(ZP))

		page := make([]byte, ZP)
		copy(page, content[:pageSize])
		mem.Pages[pageNum] = &Page{
			Value:  page,
			Access: access,
		}
		content = content[pageSize:]
	}
}

func allocateStack(mem *Memory, start, end uint32) {
	for addr := start; addr < end; addr += ZP {
		pageNum := addr / ZP
		mem.Pages[pageNum] = &Page{
			Value:  make([]byte, ZP), // fill with 0
			Access: MemoryReadWrite,
		}
	}
}

func DecodeSerializedValues(p []byte) ([]byte, []byte, []byte, uint16, uint32, error) {
	var err error

	var oLen, wLen, cLen uint64
	oLen, p, err = ReadUintFixed(p, 3)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}
	wLen, p, err = ReadUintFixed(p, 3)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}

	var z uint64
	var s uint64
	z, p, err = ReadUintFixed(p, 2)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}
	s, p, err = ReadUintFixed(p, 3)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}

	var o, w, c []byte
	o, p, err = ReadBytes(p, oLen)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}

	w, p, err = ReadBytes(p, wLen)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}

	cLen, p, err = ReadUintFixed(p, 4)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}

	c, p, err = ReadBytes(p, cLen)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}

	return c, o, w, uint16(z), uint32(s), nil
}

func ReadUintFixed(data []byte, numBytes int) (uint64, []byte, error) {
	if numBytes == 0 {
		return 0, data, nil
	}
	if numBytes > 8 || numBytes < 0 {
		return 0, data, fmt.Errorf("invalid number of octets to read")
	}
	if len(data) < numBytes {
		return 0, data, fmt.Errorf("not enough data to read a uint")
	}

	var result uint64
	for i := 0; i < numBytes; i++ {
		// little-endian
		result |= uint64(data[i]) << (8 * i)
	}

	return result, data[numBytes:], nil
}

func ReadBytes(data []byte, numBytes uint64) ([]byte, []byte, error) {
	if uint64(len(data)) < numBytes {
		return nil, data, fmt.Errorf("not enough data to read %d bytes", numBytes)
	}

	return data[:numBytes], data[numBytes:], nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
