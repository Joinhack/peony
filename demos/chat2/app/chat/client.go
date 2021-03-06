package chat

import (
	"code.google.com/p/go.net/websocket"
	"github.com/joinhack/pmsg"
	"io"
	"net"
)

type ChatClient struct {
	net.Conn
	isKickoff  bool
	clientId   uint32
	clientType byte
	wchan      chan pmsg.Msg
}

func NewChatClient(conn net.Conn, clientId uint32, clientType byte) *ChatClient {
	return &ChatClient{
		Conn:       conn,
		isKickoff:  false,
		clientId:   clientId,
		clientType: clientType,
		wchan:      make(chan pmsg.Msg, 1),
	}
}

func (conn *ChatClient) IsKickoff() bool {
	return conn.isKickoff
}

func (conn *ChatClient) Redirect(hubid int) {
	conn.SendMsg(&RedirectMsg{Location: hubAddrs[hubid]})
}

func (conn *ChatClient) Kickoff() {
	conn.isKickoff = true
	defer func() {
		if err := recover(); err != nil {
			ERROR.Println(err)
		}
	}()
	if _, err := conn.Write(KickoffJsonBytes); err != nil {
		ERROR.Println(err)
	}
	conn.Conn.Close()
}

func (conn *ChatClient) Id() uint32 {
	return conn.clientId
}

func (conn *ChatClient) Type() byte {
	return conn.clientType
}

func (conn *ChatClient) SendMsg(msg pmsg.Msg) {
	if msg == nil {
		return
	}
	conn.wchan <- msg
}

func (conn *ChatClient) CloseChannel() {
	defer func() {
		recover()
	}()
	close(conn.wchan)
}

func (conn *ChatClient) SendMsgLoop() {
	var msg pmsg.Msg
	var ok bool
	var err error
	defer func() {
		if err := recover(); err != nil {
			ERROR.Println(err)
		}
		conn.Conn.Close()
	}()

	for {
		select {
		case msg, ok = <-conn.wchan:
			if !ok {
				// the channel is closed
				return
			}
			if _, err = conn.Write(msg.Body()); err != nil {
				conn.CloseChannel()
				if err == io.EOF {
					ws := conn.Conn.(*websocket.Conn)
					INFO.Println(ws.Request().RemoteAddr, "closed")
				} else {
					ERROR.Println(err)
				}
				//the receive task will exit.
				return
			}
		}
	}
}
