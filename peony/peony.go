package main

import (
	"flag"
	"os"
)

var usage = `peony apppath`

var (
	apppath *string = flag.String("apppath", "", "the path of app")
)

func main() {
	flag.Parse()
	if *apppath == "" {
		println(usgae)
		os.exit(1)
	}

}
