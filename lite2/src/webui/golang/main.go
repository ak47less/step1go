package main

import (
	"github.com/ak47less/step1go/lite2/app"
	"github.com/ak47less/step1go/lite2/lib"
)

func main() {

	a := new(app.Step1Lite2App)
	ws := new(lib.WebServerConfig)

	ws.Enabled = true
	ws.Host = ""
	ws.Port = 9853

	// ws.Port = 80

	cfg := a.Configure()
	cfg.WebServer = ws

	err := a.Run()
	if err != nil {
		panic(err)
	}

}
