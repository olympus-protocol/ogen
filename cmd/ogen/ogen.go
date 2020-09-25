package main

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/commands"
	_ "net/http/pprof"
)

func main() {
	err := commands.Execute()
	if err != nil {
		panic(err)
	}
}
