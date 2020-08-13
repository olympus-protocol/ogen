package main

import (
	"github.com/olympus-protocol/ogen/internal/cli"
	_ "net/http/pprof"
)

func main() {
	cli.Execute()
}
