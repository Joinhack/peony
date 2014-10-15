package controllers

import (
	"code.google.com/p/go.net/websocket"
	"github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/chat2/app/chat"
)

type WebSocket struct {
}

//@Mapper("/echo", method="WS")
func (c *WebSocket) Echo(ws *websocket.Conn) {
	var bs [1024]byte
	for {
		if n, err := ws.Read(bs[:]); err != nil {
			peony.ERROR.Println(err)
			return
		} else {
			peony.INFO.Println("recv info:", string(bs[:n]))
			ws.Write(bs[:n])
		}
	}
}

//@Mapper("/", method="WS")
func (c *WebSocket) Index(ws *websocket.Conn) {
	chat.Chat(ws)
}

//@Mapper("/chat", method="WS")
func (c *WebSocket) Chat(ws *websocket.Conn) {
	chat.Chat(ws)
}
