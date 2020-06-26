package main

import (
	"github.com/olympus-protocol/ogen/cli"
	_ "net/http/pprof"
)

func main() {
	cli.Execute()
}
