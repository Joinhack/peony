package controllers

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/joinhack/peony"
	"runtime"
	"sync"
	"time"
)

type WebSocket struct {
}

func init() {
	go func() {
		for {
			fmt.Println("goroutine number", runtime.NumGoroutine())
			time.Sleep(60 * time.Second)
		}
	}()
}

var cl sync.Mutex

var clients = map[int64]*Client{}

type Msg struct {
	Code    int //0 success, -1 error
	Type    int //0 Login, 1 reply, 2 msg
	From    int64
	To      int64
	Content string
}

type Client struct {
	ws      *websocket.Conn
	msgChan chan *Msg
}

//@Mapper("/socket", method="WS")
func (c WebSocket) ChatSocket(ws *websocket.Conn) {
	ws.SetReadDeadline(time.Now().Add(30 * time.Second))
	var msg Msg
	if err := websocket.JSON.Receive(ws, &msg); err != nil {
		peony.ERROR.Println(err)
		return
	}
	if msg.Type != 0 {
		peony.ERROR.Println("please login first")
		return
	}
	var ok = &Msg{Code: 0, From: msg.From, To: msg.From, Type: 1}
	var fail = &Msg{Code: -1, From: msg.From, To: msg.From, Type: 1}
	websocket.JSON.Send(ws, ok)
	ws.SetReadDeadline(time.Now().Add(1 * time.Hour))
	client := &Client{ws: ws, msgChan: make(chan *Msg, 10)}
	id := msg.From
	cl.Lock()
	clients[id] = client
	cl.Unlock()
	defer func() {
		cl.Lock()
		delete(clients, id)
		cl.Unlock()
	}()
	go func() {
		var msg Msg
		for {
			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				peony.ERROR.Println(err)
				close(client.msgChan)
				return
			}
			msg.From = id
			var isOk bool
			var toClient *Client
			var rs = ok
			cl.Lock()
			if toClient, isOk = clients[msg.To]; !isOk {
				rs = fail
			}
			cl.Unlock()
			if err := websocket.JSON.Send(ws, rs); err != nil {
				peony.ERROR.Println(err)
				return
			}
			if toClient != nil {
				toClient.msgChan <- &msg
			}
		}
	}()
	for {
		msg := <-client.msgChan
		if err := websocket.JSON.Send(ws, msg); err != nil {
			peony.ERROR.Println(err)
			return
		}
	}
}
