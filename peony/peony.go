package main

import (
	"flag"
)

var (
	app *string = flag.String("app", "", "the path of app")
)

func main() {
	flag.Parse()

}
