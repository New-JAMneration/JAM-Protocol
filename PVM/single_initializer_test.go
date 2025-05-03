package PolkaVM

import (
	"testing"
)

func TestSingleInitializer(t *testing.T) {
	p := []byte{
		0x05, 0x00, 0x00, // o length 5
		0x04, 0x00, 0x00, // w length 4
		0x10, 0x00, // z = 16
		0x20, 0x00, 0x00, // s = 32
		// o content (5 bytes)
		'o', 'b', 'j', '1', '0',
		// w content (4 bytes)
		'w', 'd', 'a', 't',
		// c length 7
		0x07, 0x00, 0x00, 0x00,
		// c content (7 bytes)
		'c', 'o', 'd', 'e', 'd', 'a', 't',
	}

	a := []byte("args_data")

	c, regs, mem, err := SingleInitializer(p, a)
	if err != nil {
		t.Fatalf("singleInitializer returned an error: %v", err)
	}

	// validate c
	expectedC := []byte("codedat")
	if string(c) != string(expectedC) {
		t.Errorf("Expected c to be %v, got %v", expectedC, c)
	}

	// validate registers
	if regs[0] != uint64(1<<32-1<<16) {
		t.Errorf("Unexpected register[0]: got %v", regs[0])
	}
	if regs[1] != uint64(1<<32-2*ZZ-ZI) {
		t.Errorf("Unexpected register[1]: got %v", regs[1])
	}
	if regs[7] != uint64(1<<32-ZZ-ZI) {
		t.Errorf("Unexpected register[7]: got %v", regs[7])
	}
	if regs[8] != uint64(len(a)) {
		t.Errorf("Unexpected register[8]: expected %d, got %d", len(a), regs[8])
	}

	// valiadate o in read-only memory
	readOnlyPageNum := uint32(ZZ / ZP)
	if page, exists := mem.Pages[readOnlyPageNum]; !exists {
		t.Errorf("Expected read-only memory at page %d, but not found", readOnlyPageNum)
	} else if page.Access != MemoryReadOnly {
		t.Errorf("Expected read-only access for page %d, got %v", readOnlyPageNum, page.Access)
	} else if string(page.Value[:5]) != "obj10" {
		// fmt.Println(page.Value)
		t.Errorf("Expected 'obj10' in read-only page, got %v", string(page.Value[:5]))
	}

	// validate w in read-write memory
	readWritePageNum := uint32((2*ZZ + Z(len("obj10"))) / ZP)
	if page, exists := mem.Pages[readWritePageNum]; !exists {
		t.Errorf("Expected read-write memory at page %d, but not found", readWritePageNum)
	} else if page.Access != MemoryReadWrite {
		t.Errorf("Expected read-write access for page %d, got %v", readWritePageNum, page.Access)
	} else if string(page.Value[:4]) != "wdat" {
		t.Errorf("Expected 'wdat' in read-write page, got %v", string(page.Value[:4]))
	}

	// validate a in argument memory
	stackEnd := uint32(1<<32 - 2*ZZ - ZI)
	argumentPageNum := uint32(stackEnd / ZP)
	if page, exists := mem.Pages[argumentPageNum]; !exists {
		t.Errorf("Expected argument memory at page %d, but not found", argumentPageNum)
	} else if page.Access != MemoryReadOnly {
		t.Errorf("Expected read-only access for argument page %d, got %v", argumentPageNum, page.Access)
	} else if string(page.Value[:len(a)]) != "args_data" {
		t.Errorf("Expected 'args_data' in argument page, got %v", string(page.Value[:len(a)]))
	}

	// validate stack memory
	stackStart := stackEnd - P(int(32))
	for addr := stackStart; addr < stackEnd; addr += ZP {
		pageNum := addr / ZP
		if page, exists := mem.Pages[pageNum]; !exists {
			t.Errorf("Expected stack memory at page %d, but not found", pageNum)
		} else if page.Access != MemoryReadWrite {
			t.Errorf("Expected read-write access for stack page %d, got %v", pageNum, page.Access)
		}
	}
}
