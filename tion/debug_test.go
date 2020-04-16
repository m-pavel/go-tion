package tion

import "testing"

func TestBytest1(t *testing.T) {
	ba := []byte{0, 1, 2, 3}
	v := Bytes(ba)
	if v != "{0x0, 0x1, 0x2, 0x3}" {
		t.Fatal(v)
	}
}

func TestBytest2(t *testing.T) {
	ba := []byte{10, 11, 12, 13, 14, 15, 16, 17}
	v := Bytes(ba)
	if v != "{0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11}" {
		t.Fatal(v)
	}
}
