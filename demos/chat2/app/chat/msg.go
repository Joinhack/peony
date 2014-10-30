package chat

import (
	"encoding/json"
	
	"fmt"
)

var (
	OkJsonBytes, _             = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": 0})
	LoginFailJsonBytes, _      = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "login fail"})
	LoginAlterJsonBytes, _     = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "please login"})
	UnknownDevicesJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "unknown devices"})

	UnknownMsgTypeJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "unknown message type"})

	InvaildParameters, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "invaild parameters, please check."})

	ErrorJsonFormatJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "json format error"})

	JsonFormatErrorJsonBytes, _      = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "format json error"})
	ErrorMsgIdJsonFormatJsonBytes, _ = json.Marshal(map[string]interface{}{"type": SysReplyMsgType, "code": -1, "msg": "error msg id"})

	KickoffJsonBytes, _ = json.Marshal(map[string]interface{}{"type": OffineMsgType, "code": -1, "msg": "You are kick to offline."})
)

const (
	SysReplyMsgType = 255
	LoginMsgType    = 254
	RedirectMsgType = 253
	ReplyMsgType    = 252

	NotifyMsgType = 249
	OffineMsgType = 248
	PingMsgType   = 0
	UserMsgType   = 1
	GroupMsgType  = 2

	TextMsgBodyType            = 1
	ImageMsgBodyType           = 2
	FileMsgBodyType            = 3
	SoundMsgBodyType           = 4
	StickMsgBodyType           = 5
	ReadedMsgBodyType          = 6
	LocationMsgBodyType        = 7
	GroupAddMsgBodyType        = 8
	GroupMemberDelMsgBodyType  = 9
	GroupRenameMsgBodyType     = 10
	GroupRemoveMsgBodyType     = 11
	TokenRegisterMsgBodyType   = 251
	TokenUnregisterMsgBodyType = 250
)

type LoginMsg struct {
	Id       uint32 `json:"id"`
	Password string `json:"password"`
	DevType  byte   `json:"devType"`
	Type     byte   `json:"type"`
	Time     int64  `json:"time"`
}

type UserInfo struct {
	Id       uint32 `json:"id"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Msg struct {
	Id     string     `json:"id"`
	Sender *string    `json:"sender,omitempty"`
	From   uint32     `json:"from,,omitempty"`
	To     uint32     `json:"to,omitempty"`
	Time   int64      `json:"time,omitempty"`
	Type   byte       `json:"type,omitempty"`
	Bodies *[]MsgBody `json:"bodies,omitempty"`
	Ext    *[]byte    `json:"ext,omitempty"`
}

type MsgBody interface {
	GetType() byte
}

type ContentMsgBody struct {
	Type    byte   `json:"type"`
	Content string `json:"content,omitempty"`
}

func (body *ContentMsgBody) GetType() byte {
	return body.Type
}

type ImageMsgBody struct {
	Type      byte   `json:"type"`
	ScaledUrl string `json:"surl,omitempty"`
	Url       string `json:"url,omitempty"`
	Name      string `json:"name,omitempty"`
}

func (body *ImageMsgBody) GetType() byte {
	return body.Type
}

type FileMsgBody struct {
	Type byte   `json:"type"`
	Url  string `json:"url,omitempty"`
	Size int32  `json:"filesize,omitempty"`
	Name string `json:"name,omitempty"`
}

func (body *FileMsgBody) GetType() byte {
	return body.Type
}

type SoundMsgBody struct {
	Type byte   `json:"type"`
	Url  string `json:"url,omitempty"`
	Size int32  `json:"filesize,omitempty"`
}

func (body *SoundMsgBody) GetType() byte {
	return body.Type
}

type LocationMsgBody struct {
	Type byte   `json:"type"`
	Lat  string `json:"lat,omitempty"`
	Long string `json:"lng,omitempty"`
}

func (body *LocationMsgBody) GetType() byte {
	return body.Type
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
	Type byte   `json:"type"`
	NId  string `json:"nid,omitempty"`
	Time int64  `json:"time,omitempty"`
	Code int    `json:"code"`
	Msg  string `json:"msg, omitempty"`
	Id   string `json:"id"`
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

func NewReplySuccessMsg(id uint32, msgid string, t int64) *ReplyMsg {
	return &ReplyMsg{
		Type: ReplyMsgType,
		NId:  NewMsgId(id, t),
		Time: t / 1000000,
		Id:   msgid,
	}
}

func NewReplyFailMsg(id uint32, msgid string, msg string) *ReplyMsg {
	return &ReplyMsg{
		Type: ReplyMsgType,
		Code: -1,
		Id:   msgid,
		Msg:  msg,
	}
}

func NewMsgId(id uint32, t int64) string {
	return fmt.Sprintf("%d:%d", t/1000, id)
}
