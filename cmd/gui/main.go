package main

import (
	"github.com/leaanthony/mewn"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/wailsapp/wails"
)

func main() {

	go func() {
		log := config.GlobalParams.Logger

		log.Infof("Starting Ogen v%v", params.Version)
		log.Trace("Loading log on debug mode")

		config.InterruptListener()
		config.GlobalFlags.LogFile = true

		db, err := blockdb.NewLevelDB()
		if err != nil {
			log.Fatal(err)
		}

		s, err := server.NewServer(db)
		if err != nil {
			log.Fatal(err)
		}

		go s.Start()

		<-config.GlobalParams.Context.Done()

		err = s.Stop()
		if err != nil {
			log.Fatal(err)
		}
	}()

	js := mewn.String("./frontend/build/static/js/main.js")
	css := mewn.String("./frontend/build/static/css/main.css")

	app := wails.CreateApp(&wails.AppConfig{
		Width:  1024,
		Height: 768,
		Title:  "Olympus",
		JS:     js,
		CSS:    css,
		Colour: "#131313",
	})

	app.Run()
}
