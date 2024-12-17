package hash

import (
	"encoding/hex"
	"testing"
)

func TestBlake2bHash(t *testing.T) {

	hash := Blake2bHash([]byte("abc"))
	var buf = hash[:]

	output := "bddd813c634239723171ef3fee98579b94964e3bb1cb3e427262c8c068d52319"
	t.Logf("ouput: %v", hex.EncodeToString(buf))
	t.Logf("expected: %v", output)
	t.Logf("ouput: %t", output == hex.EncodeToString(buf))
}

func TestBlake2bHashPartial(t *testing.T) {
	hash := Blake2bHashPartial([]byte("abc"), 4)
	var buf = hash[:]

	output := "bddd813c"
	t.Logf("ouput: %v", hex.EncodeToString(buf))
	t.Logf("expected: %v", output)
	t.Logf("ouput: %t", output == hex.EncodeToString(buf))
}

func TestKeccakHash(t *testing.T) {

	hash2 := KeccakHash([]byte("test"))
	var buf2 = hash2[:]

	output := "9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658"
	t.Logf("expected: %v", output)
	t.Logf("ouput: %v", hex.EncodeToString(buf2))
	t.Logf("ouput: %t", output == (hex.EncodeToString(buf2)))
}
