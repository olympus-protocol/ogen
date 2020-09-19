package main

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/commands"
	_ "net/http/pprof"
)

func main() {
	commands.Execute()
}
