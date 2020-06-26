package main

import (
	_ "net/http/pprof"
	"github.com/olympus-protocol/ogen/cli"
)

func main() {
	cli.Execute()
}
