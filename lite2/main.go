package main

import "github.com/ak47less/step1go/lite2/app"

func main() {

	a := new(app.Step1Lite2App)
	err := a.Run()
	if err != nil {
		panic(err)
	}

}
