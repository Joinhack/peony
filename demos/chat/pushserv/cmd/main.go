package main

import (
	"flag"
	"github.com/joinhack/peony/demos/chat/pushserv"
	"time"
	"fmt"
)

var (
	addr *string = flag.String("addr", ":9001", "bind address")
	cfg *string = flag.String("config", "", "config path")
)

func main() {
	flag.Parse()
	pushserv, err := pushserv.NewPushServer(*addr, *cfg)
	if err != nil {
		println(err.Error())
		return
	}
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("server is in running.")
	}()
	if err := pushserv.ListenAndServe(); err != nil {
		println(err.Error())
	}
}
