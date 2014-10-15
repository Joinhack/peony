package chat

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/joinhack/pmsg"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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

func sendNotify(rmsg pmsg.RouteMsg) bool {
	if pusher != nil {
		var msg *Msg
		message := map[string]json.RawMessage{}
		err := json.Unmarshal(rmsg.Body(), &message)
		if msg, err = ConvertMsg(message); err != nil {
			ERROR.Println(err)
			return false
		}
		if err != nil {
			ERROR.Println(err)
			return false
		}

		if msg.To == 0 || msg.Bodies == nil {
			return true
		}

		var pushContent string
		var sender string
		if msg.Sender == nil {
			sender = "nobody"
		} else {
			sender = *msg.Sender
		}
		for _, body := range *msg.Bodies {
			switch body.GetType() {
			case TextMsgBodyType:
				pushContent = fmt.Sprintf("%s: %s", sender, body.(*ContentMsgBody).Content)
			case ImageMsgBodyType:
				pushContent = fmt.Sprintf("%s sent you a photo.", sender)
			case SoundMsgBodyType:
				pushContent = fmt.Sprintf("%s sent you a voice message.", sender)
			case LocationMsgBodyType:
				pushContent = fmt.Sprintf("%s sent you a location.", sender)
			case StickMsgBodyType:
				pushContent = fmt.Sprintf("%s: [sticker]", sender)
			default:
				return false
			}
		}
		token := gettokens(msg.To)
		if len(token) == 0 {
			return true
		}
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

type pushParam struct {
	dev      byte
	token    string
	contents string
}

type Pusher struct {
	wchan []chan *pushParam
	url   string
	idx   int
}

func NewPusher(n int, url string) *Pusher {
	p := &Pusher{
		wchan: make([]chan *pushParam, n),
		url:   url,
	}
	for i := 0; i < n; i++ {
		p.wchan[i] = make(chan *pushParam)
		go p.task(p.wchan[i])
	}
	return p
}

func (p *pushParam) toUrlParam() string {
	v := url.Values{}
	v.Add("dev", fmt.Sprintf("%d", p.dev))
	v.Add("token", p.token)
	v.Add("contents", p.contents)
	return v.Encode()
}

type Reply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
}

func (p *Pusher) task(channel <-chan *pushParam) {
	var ok bool
	var param *pushParam
	var err error

	client := &http.Client{}

	for {
		var req *http.Request
		var resp *http.Response
		var sli []byte
		select {
		case param, ok = <-channel:
			if !ok {
				return
			}
		}
	SEND:
		req, _ = http.NewRequest("GET", fmt.Sprintf("%s?%s", p.url, param.toUrlParam()), nil)
		req.Close = false
		resp, err = client.Do(req)
		if err != nil {
			ERROR.Println(err)
			time.Sleep(100 * time.Millisecond)
			goto SEND
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			ERROR.Println("The respose code is", resp.StatusCode)
			time.Sleep(100 * time.Millisecond)
			goto SEND
		}
		if sli, err = ioutil.ReadAll(resp.Body); err != nil {
			ERROR.Println(err)
			resp.Body.Close()
			goto SEND
		}
		resp.Body.Close()
		var reply Reply
		if err = json.Unmarshal(sli, &reply); err != nil {
			ERROR.Println(err)
			goto SEND
		}
		if reply.Code != 0 {
			ERROR.Println(reply.Msg)
		}
	}
}

func (p *Pusher) Push(dev byte, token, contents string) {
	for i := 3; i > 0; i-- {
		if p.idx >= len(p.wchan) {
			p.idx = 0
		}
		select {
		case p.wchan[p.idx] <- &pushParam{
			dev:      dev,
			token:    token,
			contents: contents}:
			p.idx++
			return
		case <-time.After(100 * time.Millisecond):
		}
		p.idx++
	}
	p.idx++
}
