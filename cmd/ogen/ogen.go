package main

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/commands"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go http.ListenAndServe(":1234", nil)
	err := commands.Execute()
	if err != nil {
		panic(err)
	}
}
