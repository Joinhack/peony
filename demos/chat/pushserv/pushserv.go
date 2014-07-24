package pushserv

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/joinhack/fqueue"
	. "github.com/joinhack/goconf"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	InvaildLength = errors.New("invalid slice length")
	UnkonwnDevice = errors.New("unkown device")
)

type PushServer struct {
	conf    Config
	queue   fqueue.Queue
	server  *http.Server
	ignores map[string]uint32
	clients []Client
}

var mode string

func (p *PushServer) customer() {
	var idx = 0
	for {
		var sli []byte
		var err error
		var ok bool
		var r Request
		var ignoreTime uint32
		if sli, err = p.queue.Pop(); err != nil {
			if err == fqueue.QueueEmpty {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			panic(err)
		}
		if err = r.unmarshal(sli); err != nil {
			log.Println(err)
			continue
		}
		if r.dev != 0 {
			log.Println("don't support now")
			continue
		}
		k := hex.EncodeToString(r.token)
		if ignoreTime, ok = p.ignores[k]; ok && r.time < ignoreTime {
			continue
		}
	SEND:
		if idx >= len(p.clients) {
			idx = 0
		}
		if err = p.clients[idx].SendRequest(&r); err != nil {
			if err == SendTimeout {
				//try other connection.
				idx++
				goto SEND
			}
		}
		idx++
	}
}

func (p *PushServer) Close() error {
	for _, client := range p.clients {
		if client != nil {
			client.Close()
		}
	}
	if p.queue != nil {
		return p.queue.Close()
	}
	return nil
}

func mergeValues(req *http.Request) url.Values {
	urlQuery := req.URL.Query()
	reqForm := req.Form
	l := len(urlQuery) + len(reqForm)
	var values url.Values
	switch l {
	case len(urlQuery):
		values = urlQuery
	case len(reqForm):
		values = reqForm
	}

	if values == nil {
		values = make(url.Values, l)
		for k, v := range urlQuery {
			values[k] = append(values[k], v...)
		}
		for k, v := range reqForm {
			values[k] = append(values[k], v...)
		}
	}
	return values
}

type Reply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
}

func (p *PushServer) sendReply(w http.ResponseWriter, reply *Reply) {
	var sli []byte
	var err error
	if sli, err = json.Marshal(reply); err != nil {
		_, err = w.Write([]byte("json format error"))
		return
	}
	if _, err = w.Write(sli); err != nil {
		log.Println(err)
	}
}

type Request struct {
	dev       byte
	time      uint32
	token     []byte
	contents  string
}

func (r *Request) unmarshal(p []byte) error {
	var l uint16
	if len(p) <= 5 {
		return InvaildLength
	}

	r.dev = p[0]
	if r.dev != 0 && r.dev != 1 {
		return UnkonwnDevice
	}
	p = p[1:]
	r.time = binary.LittleEndian.Uint32(p)
	p = p[4:]
	l = binary.LittleEndian.Uint16(p)
	p = p[2:]
	r.token = p[:l]
	p = p[l:]
	l = binary.LittleEndian.Uint16(p)
	p = p[2:]
	r.contents = string(p)
	p = p[l:]
	if len(p) != 0 {
		return InvaildLength
	}
	return nil
}

func (r *Request) Bytes() []byte {
	var n int
	var p, sli []byte
	sli = make([]byte, 1+4+2+len(r.token)+2+len(r.contents))
	p = sli
	p[0] = byte(r.dev)
	p = p[1:]
	binary.LittleEndian.PutUint32(p, uint32(r.time))
	p = p[4:]
	binary.LittleEndian.PutUint16(p, uint16(len(r.token)))
	p = p[2:]
	n = copy(p, r.token)
	p = p[n:]
	binary.LittleEndian.PutUint16(p, uint16(len(r.contents)))
	p = p[2:]
	n = copy(p, []byte(r.contents))
	p = p[n:]
	return sli
}

func (p *PushServer) push2queue(r *Request) (err error) {
	var sli []byte = r.Bytes()
	var maxTimes int = 3
	for maxTimes > 0 {
		if err = p.queue.Push(sli); err != nil {
			if err == fqueue.NoSpace {
				time.Sleep(10 * time.Millisecond)
				maxTimes--
			}
			return err
		}
		return
	}
	return
}

func (p *PushServer) Push(w http.ResponseWriter, req *http.Request) {
	var err error
	var tokenSli []byte
	params := mergeValues(req)
	var reply = Reply{Code: -1}

	dev := strings.Trim(params.Get("dev"), " ")
	if dev == "" {
		reply.Msg = "invalidate dev"
		p.sendReply(w, &reply)
		return
	}

	var devId int
	if devId, err = strconv.Atoi(dev); err != nil {
		reply.Msg = err.Error()
		p.sendReply(w, &reply)
		return
	}

	contents := strings.Trim(params.Get("contents"), " ")
	if contents == "" {
		reply.Msg = "contents can't be empty"
		p.sendReply(w, &reply)
		return
	}
	token := strings.Trim(params.Get("token"), " ")
	if token == "" {
		reply.Msg = "invalidate token"
		p.sendReply(w, &reply)
		return
	} else if tokenSli, err = hex.DecodeString(token); err != nil {
		reply.Msg = "invalidate token"
		p.sendReply(w, &reply)
		return
	}
	if err = p.push2queue(&Request{
		dev:      byte(devId),
		contents: contents,
		token:    tokenSli,
	}); err != nil {
		reply.Msg = err.Error()
		p.sendReply(w, &reply)
		return
	}
	reply.Code = 0
	reply.Msg = ""
	p.sendReply(w, &reply)
}

func (p *PushServer) ListenAndServe() error {
	return p.server.ListenAndServe()
}

func NewPushServer(addr string, cfgpath string) (pushServer *PushServer, err error) {
	conf := Config{}
	var binPath string
	if binPath, err = filepath.Abs(os.Args[0]); err != nil {
		return
	}
	path := filepath.Dir(binPath)
	if cfgpath == "" {
		conf.ReadFile(filepath.Join(path, "..", "config", "pushsvr.cnf"))
	} else {
		conf.ReadFile(cfgpath)
	}
	ps := &PushServer{ignores: map[string]uint32{}}

	ps.conf = conf
	fqueue.QueueLimit = int(conf.IntDefault(mode, "fileLimit", 1024*1024*100))
	clientNum := int(conf.IntDefault(mode, "clientNum", 10))
	apnsGW := conf.StringDefault(mode, "apnsGateWay", "gateway.sandbox.push.apple.com:2195")
	apnsCert := conf.StringDefault(mode, "apnsCert", "")
	apnsKey := conf.StringDefault(mode, "apnsKey", "")

	queueFile := conf.StringDefault(mode, "queueFile", "/tmp/pushsvr.data")
	if ps.queue, err = fqueue.NewFQueue(queueFile); err != nil {
		ps.Close()
		return
	}
	ps.clients = make([]Client, clientNum)
	for i := 0; i < clientNum; i++ {
		ps.clients[i], err = NewAPNSClient(apnsGW, apnsCert, apnsKey)
		if err != nil {
			return
		}
	}
	http.HandleFunc("/push", ps.Push)
	ps.server = &http.Server{Addr: addr}
	pushServer = ps
	go pushServer.customer()
	return
}
