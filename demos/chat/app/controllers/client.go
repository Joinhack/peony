package controllers

import (
	"github.com/joinhack/peony"
	"github.com/joinhack/pmsg"
	"net"
	"sync"
)

type ChatClient struct {
	net.Conn
	mutex      *sync.Mutex
	isKickoff  bool
	clientId   uint64
	clientType byte
	wchan      chan pmsg.Msg
}

func NewChatClient(conn net.Conn, clientId uint64, clientType byte, mutex *sync.Mutex) *ChatClient {
	return &ChatClient{
		Conn:       conn,
		isKickoff:  false,
		clientId:   clientId,
		clientType: clientType,
		mutex:      mutex,
		wchan:      make(chan pmsg.Msg, 1),
	}
}

func (conn *ChatClient) IsKickoff() bool {
	return conn.isKickoff
}

func (conn *ChatClient) Kickoff() {
	conn.isKickoff = true
	defer func() {
		if err := recover(); err != nil {
			peony.ERROR.Println(err)
		}
	}()
	//TODO: send kickoff msg
	conn.Conn.Close()
}

func (conn *ChatClient) Id() uint64 {
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

func (conn *ChatClient) SendMsgLoop() {
	defer func() {
		if err := recover(); err != nil {
			peony.ERROR.Println(err)
		}
	}()
	var msg pmsg.Msg
	var ok bool
	var err error
	for {
		select {
		case msg, ok = <-conn.wchan:
			if !ok {
				// the channel is closed
				return
			}
			conn.mutex.Lock()
			_, err = conn.Write(msg.Body())
			conn.mutex.Unlock()
			if err != nil {
				peony.ERROR.Println(err)
				conn.Conn.Close()
				return
			}
		}

	}
}
