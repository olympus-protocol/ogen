package main

import (
	"github.com/olympus-protocol/ogen/cli"
	"github.com/olympus-protocol/ogen/logger"
	"os"
)

var log = logger.New(os.Stdin)

func main() {
	cli.Execute()
}


