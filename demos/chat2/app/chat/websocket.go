package chat

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"errors"
	"github.com/joinhack/pmsg"
	"io"
	"time"
)

var (
	MustbeGroupMsg   = errors.New("Message must be group message")
	UnknowJsonFormat = errors.New("unknow json format")
)

func sendMsg(msg *Msg, msgType byte) error {
	bs, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	rmsg := pmsg.NewDeliverMsg(msgType, uint32(msg.To), bs)
	hub.Dispatch(rmsg)
	return nil
}

func sendGroupMsg(msg *Msg, msgType byte) error {
	if msg.Type != 2 {
		return MustbeGroupMsg
	}
	members := getGroupMembers(msg.To)
	for _, member := range members {
		if member == uint64(msg.From) {
			continue
		}
		msg.To = uint32(member)
		sendMsg(msg, msgType)
	}
	return nil
}

func convertBodies(bs json.RawMessage) (bodies []MsgBody, err error) {
	mbs := []map[string]json.RawMessage{}
	if err = json.Unmarshal(bs, &mbs); err != nil {
		return
	}
	for _, body := range mbs {
		var bodyType byte
		if err = ParseField(body, "type", &bodyType); err != nil {
			ERROR.Println("parse type error.")
			return
		}
		switch bodyType {
		case TextMsgBodyType, StickMsgBodyType:
			var content string
			if err = ParseField(body, "content", &content); err != nil {
				ERROR.Println("parse content error.")
				return
			}
			bodies = append(bodies, &ContentMsgBody{
				Type:    bodyType,
				Content: content,
			})
		case ImageMsgBodyType:
			var url, surl, name string
			if err = ParseField(body, "url", &url); err != nil {
				ERROR.Println("parse url error.")
				return
			}
			if err = ParseField(body, "surl", &surl); err != nil {
				ERROR.Println("parse surl error.")
				return
			}
			if err = ParseField(body, "name", &name); err != nil {
				ERROR.Println("parse name error.")
				return
			}
			bodies = append(bodies, &ImageMsgBody{
				Type:      bodyType,
				ScaledUrl: surl,
				Url:       url,
				Name:      name,
			})
		case FileMsgBodyType:
			var url, name string
			if err = ParseField(body, "url", &url); err != nil {
				ERROR.Println("parse url error.")
				return
			}
			if err = ParseField(body, "name", &name); err != nil {
				ERROR.Println("parse name error.")
				return
			}
			bodies = append(bodies, &FileMsgBody{
				Type: bodyType,
				Url:  url,
				Name: name,
			})
		case LocationMsgBodyType:
			var lat, lng string
			if err = ParseField(body, "lat", &lat); err != nil {
				ERROR.Println("parse lat error.")
				return
			}
			if err = ParseField(body, "lng", &lng); err != nil {
				ERROR.Println("parse lng error.")
				return
			}
			bodies = append(bodies, &LocationMsgBody{
				Type: bodyType,
				Lat:  lat,
				Long: lng,
			})
		default:
			err = UnknowJsonFormat
			ERROR.Println("unkown type")
			return
		}
	}
	err = nil
	return
}

func ConvertMsg(m map[string]json.RawMessage) (msg *Msg, err error) {
	err = UnknowJsonFormat
	//	var msgType byte
	var to uint32
	if err = ParseField(m, "to", &to); err != nil {
		ERROR.Println("parse to error.")
		return
	}
	var ext json.RawMessage
	ext = m["ext"]
	//limit 1k for ext
	if len(ext) > 1024*1024*1 {
		ERROR.Println("extend is too large.")
		return
	}
	bodiesMessage := m["bodies"]
	var bodies []MsgBody
	if bodies, err = convertBodies(bodiesMessage); err != nil {
		return
	}

	msg = &Msg{
		To: to,
	}
	if len(bodiesMessage) > 0 {
		msg.Bodies = &bodies
	}
	if len(ext) > 0 {
		extSli := []byte(ext)
		msg.Ext = &extSli
	}
	err = nil
	return
}

func ParseField(m map[string]json.RawMessage, field string, v interface{}) (err error) {
	err = UnknowJsonFormat
	var ok bool
	var sli []byte
	if sli, ok = m[field]; !ok {
		ERROR.Println("no field found in json, field:", field)
		return
	}
	err = json.Unmarshal(sli, v)
	return
}

func Chat(ws *websocket.Conn) {
	ws.SetReadDeadline(time.Now().Add(15 * time.Second))
	var loginMsg LoginMsg
	var user *UserInfo
	var err error
	if err = websocket.JSON.Receive(ws, &loginMsg); err != nil {
		ERROR.Println(err)
		return
	}

	if loginMsg.Type != LoginMsgType || loginMsg.Id == 0 {
		ws.Write(LoginAlterJsonBytes)
		ERROR.Println("please login first")
		return
	}

	if loginMsg.DevType != 0x1 && loginMsg.DevType != 0x2 {
		ws.Write(UnknownDevicesJsonBytes)
		ERROR.Println("unknown devices")
		return
	}
	if user, err = GetUserById(loginMsg.Id); err != nil {
		ws.Write(LoginFailJsonBytes)
		ERROR.Println(err)
		return
	}

	if user.Password != loginMsg.Password {
		ERROR.Println("login fail")
		ws.Write(LoginFailJsonBytes)
		return
	}
	if _, err = ws.Write(OkJsonBytes); err != nil {
		ERROR.Println(err)
		return
	}

	client := NewChatClient(ws, user.Id, loginMsg.DevType)
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

	var pmsgType byte = pmsg.RouteMsgType
	for {
		message := map[string]json.RawMessage{}
		ws.SetReadDeadline(time.Now().Add(1 * time.Minute))
		if err := websocket.JSON.Receive(ws, &message); err != nil {
			if err == io.EOF {
				INFO.Println(ws.Request().RemoteAddr, "closed")
			} else {
				ERROR.Println(err)
			}
			return
		}
		var msgId string
		var msgType byte

		if err = ParseField(message, "type", &msgType); err != nil {
			ws.Write(UnknownDevicesJsonBytes)
			ERROR.Println("parse type error.", err)
			return
		}
		if err = ParseField(message, "id", &msgId); err != nil {
			ws.Write(UnknownDevicesJsonBytes)
			ERROR.Println("parse msgId error.", err)
			return
		}
		if msgType == PingMsgType {
			//ping, don't reply
			continue
		}

		var msg *Msg

		if msg, err = ConvertMsg(message); err != nil {
			ws.Write(UnknownDevicesJsonBytes)
			ERROR.Println(err)
			return
		}
		if msgId == "" {
			ws.Write(ErrorMsgIdJsonFormatJsonBytes)
			ERROR.Println("msgId is empty")
			return
		}

		msg.From = client.clientId
		if len(user.Name) > 0 {
			msg.Sender = &user.Name
		}
		now := time.Now()
		msg.Time = now.UnixNano() / 1000000
		msg.Type = msgType

		reply := NewReplySuccessMsg(client.clientId, msgId, now.UnixNano())
		msg.Id = reply.NId
		client.SendMsg(reply)

		if msg.Type == 2 {
			//clone msg
			if err = sendGroupMsg(msg, pmsgType); err != nil {
				ERROR.Println(err)
			}
		} else if msg.Type == 1 {
			if err = sendMsg(msg, pmsgType); err != nil {
				ERROR.Println(err)
			}
		} else {
			ERROR.Println("Unkown msg type:", msg.Type)
			return
		}
	}
}
