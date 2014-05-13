package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

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
		println(fmt.Sprintf("%s?%s", p.url, param.toUrlParam()))
		req, _ = http.NewRequest("GET", fmt.Sprintf("%s?%s", p.url, param.toUrlParam()), nil)
		req.Close = false
		resp, err = client.Do(req)
		if err != nil {
			ERROR.Println(err)
			time.Sleep(20 * time.Millisecond)
			goto SEND
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			ERROR.Println("The respose code is", resp.StatusCode)
			time.Sleep(20 * time.Millisecond)
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
			goto SEND
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
