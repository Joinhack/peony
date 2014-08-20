package controllers

import (
	"code.google.com/p/go.net/websocket"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/joinhack/peony"
	"github.com/joinhack/pmsg"
	"io"
	"strings"
	"time"
)

var (
	MustbeGroupMsg   = errors.New("Message must be group message")
	UnknowJsonFormat = errors.New("unknow json format")
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
	c.chat(ws)
}

//@Mapper("/chat", method="WS")
func (c *WebSocket) Chat(ws *websocket.Conn) {
	c.chat(ws)
}

func sendMsg(msg *Msg, msgType byte) error {
	bs, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	rmsg := pmsg.NewDeliverMsg(msgType, uint32(*msg.To), bs)
	hub.Dispatch(rmsg)
	return nil
}

func getGroupMembers(gid uint64) []uint64 {
	members := make([]uint64, 0, 3)
	if groupRedisPool != nil {
		var err error
		var reply []interface{}
		conn := groupRedisPool.Get()
		defer conn.Close()
		if reply, err = redis.Values(conn.Do("sdump", fmt.Sprintf("%dFT", gid))); err != nil {
			ERROR.Println(err)
			return members
		}
		if len(reply) == 0 {
			return members
		}
		var l int
		var bs []byte
		if _, err = redis.Scan(reply, &l, &bs); err != nil {
			ERROR.Println(err)
			return members
		}
		for i := 0; i < len(bs); i += l {
			var val uint64
			val = uint64(binary.LittleEndian.Uint16(bs[i:]))
			members = append(members, val)
		}
	}
	return members
}

func registerToken(id uint32, dev byte, token string) error {
	conn := tokenRedisPool.Get()
	var err error
	defer conn.Close()
	if _, err = conn.Do("set", fmt.Sprintf("tk%d", id), fmt.Sprintf("%d:%s\n", dev, token)); err != nil {
		return err
	}
	return nil
}

func unregisterToken(id uint32, dev byte, token string) error {
	conn := tokenRedisPool.Get()
	var val string
	var err error
	defer conn.Close()
	key := fmt.Sprintf("tk%d", id)
	if val, err = redis.String(conn.Do("get", key)); err != nil {
		return err
	}
	if val == "" {
		return nil
	}
	vals := strings.Split(val, "\n")
	var p = make([]string, 0, len(vals))
	item := fmt.Sprintf("%d:%s\n", dev, token)
	for _, value := range vals {
		if value != item {
			p = append(p, value)
		}
	}
	if len(p) == 0 {
		if _, err = conn.Do("del", key); err != nil {
			return err
		}
	}
	if _, err = conn.Do("set", key, strings.Join(p, "\n")); err != nil {
		return err
	}
	return nil
}

func gettokens(id uint32) []string {
	conn := tokenRedisPool.Get()
	var val string
	var err error
	defer conn.Close()
	key := fmt.Sprintf("tk%d", id)
	if val, err = redis.String(conn.Do("get", key)); err != nil {
		return []string{}
	}
	if val == "" {
		return []string{}
	}
	vals := strings.Split(val, "\n")
	return vals
}

func sendGroupMsg(msg *Msg, msgType byte) error {
	if msg.SourceType != 3 && *msg.Gid == 0 {
		return MustbeGroupMsg
	}
	members := getGroupMembers(*msg.Gid)
	for _, member := range members {
		if member == uint64(msg.From) {
			continue
		}
		m := uint32(member)
		msg.To = &m
		sendMsg(msg, msgType)
	}
	return nil
}

func validateMsg(msg *Msg) error {
	switch msg.Type {
	case TextMsgType, StickMsgType:
		if msg.Content == nil || len(*msg.Content) == 0 {
			return UnknowJsonFormat
		}
	case ImageMsgType:
		if msg.BigSrc == nil || len(*msg.BigSrc) == 0 {
			return UnknowJsonFormat
		}
		if msg.SmallSrc == nil || len(*msg.SmallSrc) == 0 {
			return UnknowJsonFormat
		}
	case FileMsgType:
		if msg.Url == nil || len(*msg.Url) == 0 {
			return UnknowJsonFormat
		}
		if msg.Name == nil || len(*msg.Name) == 0 {
			return UnknowJsonFormat
		}
	case SoundMsgType:
		if msg.Url == nil || len(*msg.Url) == 0 {
			return UnknowJsonFormat
		}
	case LocationMsgType:
		if msg.Lat == nil || len(*msg.Lat) == 0 {
			return UnknowJsonFormat
		}
		if msg.Long == nil || len(*msg.Long) == 0 {
			return UnknowJsonFormat
		}
	case GroupAddMsgType:
		if msg.Members == nil || len(*msg.Members) == 0 {
			return UnknowJsonFormat
		}
	}
	return nil
}

func (c *WebSocket) chat(ws *websocket.Conn) {
	//ws.SetReadDeadline(time.Now().Add(3 * time.Second))
	var register RegisterMsg
	var err error
	if err = websocket.JSON.Receive(ws, &register); err != nil {
		ERROR.Println(err)
		return
	}

	if register.Type != LoginMsgType || register.Id == 0 {
		ws.Write(LoginAlterJsonBytes)
		ERROR.Println("please login first")
		return
	}
	if register.DevType != 0x1 && register.DevType != 0x2 {
		ws.Write(UnknownDevicesJsonBytes)
		ERROR.Println("unknown devices")
		return
	}

	if _, err = ws.Write(OkJsonBytes); err != nil {
		ERROR.Println(err)
		return
	}

	client := NewChatClient(ws, register.Id, register.DevType)
	if err = hub.AddClient(client); err != nil {
		ERROR.Println(err)
		if redirect, ok := err.(*pmsg.RedirectError); ok {
			client.Redirect(redirect.HubId)
		}
		return
	}
	go client.SendMsgLoop()
	defer func() {
		if err := recover(); err != nil {
			ERROR.Println(err)
		}
		client.CloseChannel()
		hub.RemoveClient(client)
	}()

	var msgType byte = pmsg.RouteMsgType
	for {
		var msg Msg
		ws.SetReadDeadline(time.Now().Add(1 * time.Minute))
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			if err == io.EOF {
				INFO.Println(ws.Request().RemoteAddr, "closed")
			} else {
				ERROR.Println(err)
			}
			return
		}
		if msg.SourceType == 1 && (msg.To == nil || *msg.To == 0) {
			ws.Write(InvaildParameters)
			return
		}
		if msg.SourceType == 3 && (msg.Gid == nil || *msg.Gid == 0) {
			ws.Write(InvaildParameters)
			return
		}
		now := time.Now()
		switch msg.Type {
		case PingMsgType:
			//ping, don't reply
			continue
		case TokenRegisterMsgType:
			if msg.MsgId == "" || msg.Token == nil {
				ws.Write(ErrorMsgIdJsonFormatJsonBytes)
				return
			}
			reply := &ReplyMsg{
				Type:  ReplyMsgType,
				Time:  now.UnixNano() / 1000000,
				MsgId: msg.MsgId,
				Code:  0,
			}
			if err = registerToken(client.clientId, msg.Dev, *msg.Token); err != nil {
				reply.Code = -1
				reply.Msg = err.Error()
				ws.Write(reply.Bytes())
				return
			}
			client.SendMsg(reply)
			continue
		case TokenUnregisterMsgType:
			if msg.MsgId == "" || msg.Token == nil {
				ws.Write(ErrorMsgIdJsonFormatJsonBytes)
				return
			}
			reply := &ReplyMsg{
				Type:  ReplyMsgType,
				Time:  now.UnixNano() / 1000000,
				MsgId: msg.MsgId,
				Code:  0,
			}
			if err = unregisterToken(client.clientId, msg.Dev, *msg.Token); err != nil {
				reply.Code = -1
				reply.Msg = err.Error()
				ws.Write(reply.Bytes())
				return
			}
			client.SendMsg(reply)
			continue
		case ReadedMsgType:
			msg.From = client.clientId
		case TextMsgType, ImageMsgType, FileMsgType, SoundMsgType, StickMsgType, LocationMsgType:
			if msg.MsgId == "" {
				ws.Write(ErrorMsgIdJsonFormatJsonBytes)
				return
			}
			msg.From = client.clientId
			if len(register.Name) > 0 {
				msg.Sender = &register.Name
			}
			msg.Time = now.UnixNano() / 1000000

			reply := NewReplySuccessMsg(client.clientId, msg.MsgId, now.UnixNano())
			msg.MsgId = reply.NewMsgId
			client.SendMsg(reply)
		case GroupMemberDelMsgType, GroupRemoveMsgType, GroupAddMsgType:
			msg.From = client.clientId
			if len(register.Name) > 0 {
				msg.Sender = &register.Name
			}
			now := time.Now()
			msg.Time = now.UnixNano() / 1000000
		default:
			ERROR.Println("unknown message type, close client")
			ws.Write(UnknownMsgTypeJsonBytes)
			return
		}
		if err = validateMsg(&msg); err != nil {
			ERROR.Println(err)
			ws.Write(JsonFormatErrorJsonBytes)
			return
		}
		if msg.SourceType == 3 {
			//clone msg
			if err = sendGroupMsg(&msg, msgType); err != nil {
				ERROR.Println(err)
				ws.Write(ErrorJsonFormatJsonBytes)
			}
		} else {
			if err = sendMsg(&msg, msgType); err != nil {
				ERROR.Println(err)
				ws.Write(ErrorJsonFormatJsonBytes)
			}
		}
	}
}
