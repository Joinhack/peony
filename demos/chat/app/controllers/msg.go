package controllers

import (
	"encoding/json"
	"fmt"
)

var (
	OkJsonBytes, _             = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": 0})
	LoginAlterJsonBytes, _     = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "please login"})
	UnknownDevicesJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "unknown devices"})

	UnknownMsgTypeJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "unknown message type"})

	ErrorJsonFormatJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "json format error"})

	JsonFormatErrorJsonBytes, _      = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "format json error"})
	ErrorMsgIdJsonFormatJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "error msg id"})

	KickoffJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "You are kick off."})
)

const (
	SysReplyMsgType = 255
	LoginMsgType    = 254
	RedirectMsgType = 253
	ReplyMsgType    = 252
	PingMsgType     = 0
	TextMsgType     = 1
	ImageMsgType    = 2
	FileMsgType     = 3
	SoundMsgType    = 4
	StickMsgType    = 5
	ReadedMsgType   = 6
	LocationMsgType = 7
	GroupAddMsgType = 8
	GroupDelMsgType = 9
)

type RegisterMsg struct {
	Id      uint64 `json:"id"`
	DevType byte   `json:"devType"`
	Type    byte   `json:"type"`
	Time    int64  `json:"time"`
}

type Msg struct {
	MsgId      string    `json:"msgId"`
	From       uint64    `json:"from"`
	Type       byte      `json:"type"`
	Time       int64     `json:"time"`
	Option     int       `json:"option"`
	SourceType int       `json:"sourceType"`
	Timer      *byte     `json:"timer,omitempty"`
	To         *uint64   `json:"to,omitempty"`
	Gid        *uint64   `json:"gid,omitempty"`
	Content    *string   `json:"content,omitempty"`
	SmallSrc   *string   `json:"smallsrc,omitempty"`
	BigSrc     *string   `json:"bigsrc,omitempty"`
	Url        *string   `json:"url,omitempty"`
	Lat        *string   `json:"lat,omitempty"`
	Long       *string   `json:"long,omitempty"`
	Name       *string   `json:"name,omitempty"`
	Members    *[]uint64 `json:"members,omitempty"`
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
	Type     byte   `json:"type"`
	NewMsgId string `json:"newMsgId,omitempty"`
	Time     int64  `json:"time,omitempty"`
	Code     int    `json:"code"`
	Msg      string `json:"msg, omitempty"`
	MsgId    string `json:"msgId"`
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

func NewReplySuccessMsg(id uint64, msgid string, t int64) *ReplyMsg {
	return &ReplyMsg{
		Type:     ReplyMsgType,
		NewMsgId: NewMsgId(id, t),
		Time:     t,
		MsgId:    msgid,
	}
}

func NewReplyFailMsg(id uint64, msgid string, msg string) *ReplyMsg {
	return &ReplyMsg{
		Type:  ReplyMsgType,
		Code:  -1,
		MsgId: msgid,
		Msg:   msg,
	}
}

func NewMsgId(id uint64, t int64) string {
	return fmt.Sprintf("%d:%d", t, id)
}
