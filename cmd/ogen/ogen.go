package main

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/cli"
	_ "net/http/pprof"
)

func main() {
	cli.Execute()
}
