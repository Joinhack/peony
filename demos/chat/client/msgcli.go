package main

import (
	"code.google.com/p/go.net/websocket"
	c "github.com/joinhack/peony/demos/chat/app/controllers"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

var (
	host    *string = flag.String("host", "127.0.0.1", "destination hosts")
	port    *int    = flag.Int("port", 80, "the port")
	delay   *int    = flag.Int("delay", 5, "delay")
	start   *int    = flag.Int("start", 1, "start")
	uri     *string = flag.String("uri", "/", "the port")
	connNum *int    = flag.Int("nums", 2, "the connection numbers")
)

type SumInfo struct {
	SendInfo map[int64]int64
}

var (
	TotalRecv int64
	TotalSend int64
	running   bool
)

func dumpSums() {
	fmt.Println("total send", TotalSend, "total recv", TotalRecv)
}

func main() {
	flag.Parse()
	running = true
	target := fmt.Sprintf("ws://%s:%d%s", *host, *port, *uri)
	origin := fmt.Sprintf("http://%s", *host)
	peers := make(map[int]*websocket.Conn, *connNum)
	signalChan := make(chan os.Signal, 1)
	var replyCount int = 0
	var delayN int = *delay
	go func() {
		for {
			dumpSums()
			time.Sleep(5 * time.Second)
		}
	}()
	var startNum = *start
	var endNum = *start + *connNum
	for i := startNum; i < endNum; i++ {
		var login = &c.RegisterMsg{DevType: 1, Id: uint64(i), Type: c.LoginMsgType}
		ws, err := websocket.Dial(target+fmt.Sprintf("?user=[%d]", i), "", origin)
		if err != nil {
			panic(err)
		}
		peers[i] = ws
		err = websocket.JSON.Send(ws, login)
		if err != nil {
			fmt.Println(err)
			return
		}
		rs := &c.Msg{}
		if err = websocket.JSON.Receive(ws, rs); err != nil {
			panic(err)
		}
		replyCount++;

		go func() {
			for {
				rs := &c.Msg{}
				err = websocket.JSON.Receive(ws, rs)
				if err != nil {
					fmt.Println(err)
					return
				}
				if rs.Type == 1 {
					TotalRecv++
				}
			}
		}()

		go func(i int) {
			var msg c.Msg
			content := fmt.Sprintf("%d said: 12121212asdasd", i)
			msg.Content = &content
			for replyCount < *connNum {
				time.Sleep(100000 * time.Microsecond)
			}
			for running {
				for {
					to := uint64(rand.Intn(*connNum) + startNum)
					if to != uint64(i) {
						msg.To = &to
						break
					}
				}
				msg.Type = 1
				msg.MsgId = fmt.Sprintf("%d", msg.To);
				err := websocket.JSON.Send(ws, &msg)
				if err != nil {
					fmt.Println(err)
					return
				}
				TotalSend++
				time.Sleep(time.Duration(200*delayN) * time.Microsecond)
			}
		}(i)
	}
	fmt.Println("created connections:", *connNum)
	signal.Notify(signalChan)
	<-signalChan
	running = false
	time.Sleep(10 * time.Second)
	var con *websocket.Conn
	for _, con = range peers {
		con.Close()
	}
	dumpSums()
}
