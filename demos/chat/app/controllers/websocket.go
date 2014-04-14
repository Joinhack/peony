package controllers

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/joinhack/peony"
	"github.com/joinhack/pmsg"
	"io"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type WebSocket struct {
}

func init() {
	go func() {
		for {
			var memstats runtime.MemStats
			runtime.ReadMemStats(&memstats)
			fmt.Printf(
				`goroutine number %d
Alloc %d, Sys %d, Frees %d
HeapAlloc %d, HeapSys %d, HeapInuse %d
`, runtime.NumGoroutine(),
				memstats.Alloc, memstats.Sys, memstats.TotalAlloc,
				memstats.HeapAlloc, memstats.HeapSys, memstats.HeapInuse)
			time.Sleep(30 * time.Second)
		}
	}()
}

var (
	hub *pmsg.MsgHub

	TRACE *log.Logger

	ERROR *log.Logger

	INFO *log.Logger

	WARN *log.Logger

	hubAddrs = map[int]string{}
)

func hookLog() {
	pmsg.ERROR = peony.ERROR
	pmsg.WARN = peony.WARN
	pmsg.INFO = peony.INFO
	pmsg.TRACE = peony.TRACE

	ERROR = peony.ERROR
	WARN = peony.WARN
	INFO = peony.INFO
	TRACE = peony.TRACE
}

func init() {
	peony.OnServerInit(func(s *peony.Server) {
		clusterCfg := s.App.GetStringConfig("cluster", "")
		whoami := s.App.GetStringConfig("whoami", "")
		offlineRange := s.App.GetStringConfig("offlineRange", "")
		offlineStorePath := s.App.GetStringConfig("offlineStorePath", "")
		if clusterCfg == "" || whoami == "" || offlineRange == "" {
			panic("please set cluster")
		}
		clusterMap := map[string]string{}
		clusters := strings.Split(clusterCfg, ",")
		//hook log
		hookLog()
		cfg := &pmsg.MsgHubConfig{}
		for _, v := range clusters {
			kv := strings.Split(v, "->")
			if len(kv) != 2 {
				continue
			}
			if kv[0] == whoami {
				if i, err := strconv.Atoi(whoami); err != nil {
					panic(err)
				} else {
					cfg.Id = i
					cfg.MaxRange = 1024 * 1024
					cfg.ServAddr = kv[1]

				}

			} else {
				clusterMap[kv[0]] = kv[1]
			}
		}
		rangeStr := strings.Split(offlineRange, "-")
		if i, err := strconv.Atoi(rangeStr[0]); err != nil {
			panic(err)
		} else {
			cfg.OfflineRangeStart = uint64(i)
		}

		if i, err := strconv.Atoi(rangeStr[1]); err != nil {
			panic(err)
		} else {
			cfg.OfflineRangeEnd = uint64(i)
		}
		cfg.OfflinePath = offlineStorePath
		hub = pmsg.NewMsgHub(cfg)
		for k, v := range clusterMap {
			var i int
			var err error
			if i, err = strconv.Atoi(k); err != nil {
				panic(err)
			}
			hub.AddOutgoing(i, v)
			hubAddrs[i] = v
		}
		go hub.ListenAndServe()
		//
	})
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

	var msg Msg
	for {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			if err == io.EOF {
				INFO.Println(ws.Request().RemoteAddr, "closed")
			} else {
				ERROR.Println(err)
			}
			return
		}
		if msg.Type == 0 {
			//ping, don't reply
			continue
		}
		if msg.MsgId == "" {
			ws.Write(ErrorMsgIdJsonFormatJsonBytes)
			return
		}
		msg.From = client.clientId
		now := time.Now()
		msg.Time = now.UnixNano() / 1000000
		reply := NewReplySuccessMsg(client.clientId, msg.MsgId, msg.Time)
		client.SendMsg(reply)

		bs, err := json.Marshal(&msg)
		if err != nil {
			ERROR.Println(err)
			ws.Write(ErrorJsonFormatJsonBytes)
			return
		}
		rmsg := &pmsg.DeliverMsg{To: uint64(msg.To), Carry: bs, MsgType: pmsg.RouteMsgType}
		hub.Dispatch(rmsg)
	}
}
