package main

import (
	"log"

	"github.com/m-pavel/go-tion/impl"
)

func main() {
	if err := impl.HciInit(); err != nil {
		log.Panic(err)
	}
}
