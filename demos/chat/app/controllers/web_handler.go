package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/joinhack/peony"
	"github.com/joinhack/pmsg"
	"strconv"
	"strings"
	"time"
)

func sendNotify(rmsg pmsg.RouteMsg) bool {
	if pusher != nil {
		var msg Msg
		err := json.Unmarshal(rmsg.Body(), &msg)
		if err != nil {
			ERROR.Println(err)
			return false
		}
		if msg.To == nil {
			return true
		}

		var pushContent string
		var sender string
		if msg.Sender == nil {
			sender = "nobody"
		} else {
			sender = *msg.Sender
		}
		if msg.Option == 1 {
			pushContent = fmt.Sprintf("%s sent you a whisper message.", sender)
		} else {
			switch msg.Type {
			case TextMsgType:
				if msg.Content == nil {
					return false
				}
				pushContent = fmt.Sprintf("%s: %s", sender, *msg.Content)
			case ImageMsgType:
				pushContent = fmt.Sprintf("%s sent you a photo.", sender)
			case SoundMsgType:
				pushContent = fmt.Sprintf("%s sent you a voice message.", sender)
			case LocationMsgType:
				pushContent = fmt.Sprintf("%s sent you a location.", sender)
			case StickMsgType:
				pushContent = fmt.Sprintf("%s: [sticker]", sender)
			case NotifyMsgType:
				if msg.Content == nil || len(*msg.Content) == 0 {
					return false
				}
				pushContent = *msg.Content
			default:
				return false
			}
		}

		token := gettokens(*msg.To)
		for _, tk := range token {
			if tk == "" {
				continue
			}
			tks := strings.Split(tk, ":")
			if len(tks) != 2 {
				ERROR.Println("unkonwn token", tk)
				continue
			}
			var dev int
			var err error
			if dev, err = strconv.Atoi(tks[0]); err != nil {
				ERROR.Println("unkonwn token", tk)
				continue
			}

			pusher.Push(byte(dev), tks[1], pushContent)
		}
	}
	return true
}

type WebHandler struct {
}

//@Mapper("/notify", methods=["POST", "GET"])
func (h *WebHandler) Notify(to uint32, msg string, pushContent string) peony.Renderer {
	if len(msg) == 0 || to == 0 {
		return peony.RenderJson(map[string]interface{}{"code": -1, "msg": "invalid parameters."})
	}
	now := time.Now()
	raw := make(map[string]interface{})
	if err := json.Unmarshal([]byte(msg), &raw); err != nil {
		ERROR.Println(err)
		return peony.RenderJson(map[string]interface{}{"code": -1, "msg": "message must be json"})
	}
	message := &Msg{From: 0, MsgId: "nil", Type: NotifyMsgType, Raw: raw, Time: now.UnixNano() / 1000000, To: &to}
	if len(pushContent) > 0 {
		message.Content = &pushContent
	}
	if err := sendMsg(message, pmsg.RouteMsgType); err != nil {
		return peony.RenderJson(map[string]interface{}{"code": -1, "msg": err.Error()})
	}
	return peony.RenderJson(map[string]interface{}{"code": 0})
}

var invalidParam = map[string]interface{}{"code": -1, "msg": "invalid parameters."}

//@Mapper("/group/event", methods=["POST", "GET"])
func (h *WebHandler) GroupEvent(from, to uint32, gid uint64, members, name string, event int) peony.Renderer {
	if from == 0 || gid == 0 {
		return peony.RenderJson(invalidParam)
	}
	var msgType byte = GroupAddMsgType
	if event == 1 {
		msgType = GroupMemberDelMsgType
	} else if event == 2 {
		msgType = GroupRenameMsgType
	} else if event == 3 {
		msgType = GroupRemoveMsgType
	}
	message := &Msg{
		From:       from,
		MsgId:      "nil",
		Type:       msgType,
		Gid:        &gid,
		SourceType: 3,
	}
	if msgType == GroupAddMsgType || msgType == GroupMemberDelMsgType || msgType == GroupRemoveMsgType {
		if len(members) == 0 {
			return peony.RenderJson(invalidParam)
		}
		if msgType == GroupAddMsgType && len(name) != 0 {
			message.Name = &name
		}
		mSli := strings.Split(members, ",")

		memberSli := make([]uint32, 0, len(mSli))
		for _, m := range mSli {
			if i, err := strconv.ParseUint(m, 10, 0); err != nil {
				return peony.RenderJson(map[string]interface{}{"code": -1, "msg": "invalid members."})
			} else {
				memberSli = append(memberSli, uint32(i))
			}
		}
		message.Members = &memberSli
	} else {
		if len(name) == 0 {
			return peony.RenderJson(invalidParam)
		}
		message.Name = &name
	}

	var err error
	if msgType == GroupRemoveMsgType && to != 0 {
		message.To = &to
		err = sendMsg(message, pmsg.RouteMsgType)
	} else {
		err = sendGroupMsg(message, pmsg.RouteMsgType)
	}
	if err != nil {
		return peony.RenderJson(map[string]interface{}{"code": -1, "msg": err.Error()})
	}
	return peony.RenderJson(map[string]interface{}{"code": 0})
}
