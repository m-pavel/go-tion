package tion

import "fmt"

func Bytes(in []byte) string {
	res := ""
	for i := 0; i < len(in)-1; i++ {
		res += fmt.Sprintf("0x%x, ", in[i])
	}
	res += fmt.Sprintf("0x%x", in[len(in)-1])
	return res
}
