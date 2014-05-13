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
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	MustbeGroupMsg   = errors.New("Message must be group message")
	UnknowJsonFormat = errors.New("unknow json format")
)

type WebSocket struct {
}

var statusFmt = `goroutine number %d
Alloc %d, Sys %d, Frees %d
HeapAlloc %d, HeapSys %d, HeapInuse %d
`

func init() {
	go func() {
		for {
			var memstats runtime.MemStats
			runtime.ReadMemStats(&memstats)
			fmt.Printf(statusFmt,
					runtime.NumGoroutine(),
					memstats.Alloc, memstats.Sys, memstats.TotalAlloc,
					memstats.HeapAlloc, memstats.HeapSys, memstats.HeapInuse)
			time.Sleep(30 * time.Second)
		}
	}()
}

var (
	hub *pmsg.MsgHub

	pusher *Pusher

	groupRedisPool *redis.Pool

	tokenRedisPool *redis.Pool

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

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func init() {
	peony.OnServerInit(func(s *peony.Server) {
		clusterCfg := s.App.GetStringConfig("cluster", "")
		pushsvr := s.App.GetStringConfig("push.url", "")
		pushnum := s.App.GetStringConfig("push.num", "")
		groupServer := s.App.GetStringConfig("group.server", "")
		groupPasswd := s.App.GetStringConfig("group.passwd", "")

		tokenServer := s.App.GetStringConfig("token.server", "")
		tokenPasswd := s.App.GetStringConfig("token.passwd", "")
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

		if groupServer != "" {
			groupRedisPool = newPool(groupServer, groupPasswd)
			rconn := groupRedisPool.Get()
			if _, err := rconn.Do("ping"); err != nil {
				panic(err)
			}
			rconn.Close()
		}

		if tokenServer != "" {
			tokenRedisPool = newPool(tokenServer, tokenPasswd)
			rconn := tokenRedisPool.Get()
			if _, err := rconn.Do("ping"); err != nil {
				panic(err)
			}
			rconn.Close()
		}

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

		if pushsvr != "" && pushnum != "" {
			var num int
			if i, err := strconv.Atoi(pushnum); err != nil {
				panic(err)
			} else {
				num = i
			}
			pusher = NewPusher(num, pushsvr)
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
		hub.AddOfflineMsgFilter(sendNotify)
		go hub.ListenAndServe()
	})
}

func sendNotify(rmsg pmsg.RouteMsg) bool {
	if pusher != nil {
		var msg Msg
		err := json.Unmarshal(rmsg.Body(), &msg)
		if err != nil {
			ERROR.Println(err)
			return false
		}
		if msg.To == nil || msg.Content == nil {
			return true
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
			pusher.Push(byte(dev), tks[1], *msg.Content)
		}
	}
	return true
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
	rmsg := &pmsg.DeliverMsg{To: uint64(*msg.To), Carry: bs, MsgType: msgType}
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

func registerToken(id uint64, dev byte, token string) error {
	conn := tokenRedisPool.Get()
	var err error
	defer conn.Close()
	if _, err = conn.Do("set", fmt.Sprintf("tk%d", id), fmt.Sprintf("%d:%s\n", dev, token)); err != nil {
		return err
	}
	return nil
}

func unregisterToken(id uint64, dev byte, token string) error {
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

func gettokens(id uint64) []string {
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
		if member == msg.From {
			continue
		}
		msg.To = &member
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
		ws.SetReadDeadline(time.Now().Add(6 * time.Minute))
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			if err == io.EOF {
				INFO.Println(ws.Request().RemoteAddr, "closed")
			} else {
				ERROR.Println(err)
			}
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
		case TextMsgType, ImageMsgType, FileMsgType, SoundMsgType, StickMsgType, LocationMsgType:
			if msg.MsgId == "" {
				ws.Write(ErrorMsgIdJsonFormatJsonBytes)
				return
			}
			msg.From = client.clientId
			msg.Time = now.UnixNano() / 1000000

			reply := NewReplySuccessMsg(client.clientId, msg.MsgId, msg.Time)
			msg.MsgId = reply.NewMsgId
			client.SendMsg(reply)
		case GroupDelMsgType, GroupAddMsgType:
			msg.From = client.clientId
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
