package controllers

import (
	"encoding/json"
	"fmt"
	"time"
)

var (
	OkJsonBytes, _             = json.Marshal(map[string]interface{}{"type": 0, "code": 0})
	LoginAlterJsonBytes, _     = json.Marshal(map[string]interface{}{"type": 0, "code": -1, "msg": "please login"})
	UnknownDevicesJsonBytes, _ = json.Marshal(map[string]interface{}{"type": 0, "code": -1, "msg": "unknown devices"})

	ErrorJsonFormatJsonBytes, _ = json.Marshal(map[string]interface{}{"type": 0, "code": -1, "msg": "json format error"})

	JsonFormatErrorJsonBytes, _      = json.Marshal(map[string]interface{}{"type": 0, "code": -1, "msg": "format json error"})
	ErrorMsgIdJsonFormatJsonBytes, _ = json.Marshal(map[string]interface{}{"type": 0, "code": -1, "msg": "error msg id"})

	KickoffJsonBytes, _ = json.Marshal(map[string]interface{}{"type": 0, "code": -1, "msg": "You are kick off."})
)

const (
	RedirectMsgType = 255
)

type RegisterMsg struct {
	Id      uint64 `json:"id"`
	DevType byte   `json:"devType"`
	Type    byte   `json:"type"`
}

type Msg struct {
	MsgId   string `json:"msgId"`
	From    uint64 `json:"from"`
	To      uint64 `json:"to"`
	Type    byte   `json:"type"`
	Content string `json:"content"`
	Gid     uint64 `json:"gid"`
}

type RedirectMsg struct {
	Type     byte   `json:"type"`
	Location string `json:"location"`
}

func NewRedirectMsg(location string) *RedirectMsg {
	return &RedirectMsg{Type: RedirectMsgType, Location: location}
}

func (msg *RedirectMsg) Body() []byte {
	if bs, err := json.Marshal(msg); err != nil {
		ERROR.Println(err)
		return JsonFormatErrorJsonBytes
	} else {
		return bs
	}
}

func (msg *RedirectMsg) Bytes() []byte {
	return msg.Body()
}

type ReplyMsg struct {
	Type  byte   `json:"type"`
	NewId string `json:"newId"`
	MsgId string `json:"msgId"`
}

func (msg *ReplyMsg) Body() []byte {
	if bs, err := json.Marshal(msg); err != nil {
		ERROR.Println(err)
		return JsonFormatErrorJsonBytes
	} else {
		return bs
	}
}

func (msg *ReplyMsg) Bytes() []byte {
	return msg.Body()
}

func NewReplyMsg(id uint64, msgid string) *ReplyMsg {
	return &ReplyMsg{
		Type:  0,
		NewId: NewMsgId(id),
		MsgId: msgid,
	}
}

func NewMsgId(id uint64) string {
	return fmt.Sprintf("%d:%d", time.Now().Unix(), id)
}
