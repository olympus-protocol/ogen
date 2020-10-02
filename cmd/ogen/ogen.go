package main

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/commands"
)

func main() {
	err := commands.Execute()
	if err != nil {
		panic(err)
	}
}
