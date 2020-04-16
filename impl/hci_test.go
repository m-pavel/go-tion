package impl

import "testing"

func Test1(t *testing.T) {
	if err := HciInit(); err != nil {
		t.Fatal(err)
	}
}
