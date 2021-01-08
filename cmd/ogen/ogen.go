package main

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/commands"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	err := commands.Execute()
	if err != nil {
		panic(err)
	}
}
