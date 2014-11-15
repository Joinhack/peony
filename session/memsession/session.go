package memsession

import (
	"container/list"
	"github.com/joinhack/peony"
	"github.com/streadway/simpleuuid"
	"net/http"
	"sync"
	"time"
)

var (
	CookieHttpOnly bool
	CookieSecure   bool
	SessionTimeout int
)

type item struct {
	elem *list.Element
	wrap *sessionWrap
}

type sessionWrap struct {
	*peony.Session
	lastAccess int64
}

func init() {
	peony.OnServerInit(func(s *peony.Server) {
		sm := &MemSessionManager{}
		sm.list = list.New()
		sm.sessions = make(map[string]*item)
		go func(sm *MemSessionManager) {
			sm.clearTask()
		}(sm)
		s.RegisterSessionManager(sm)
		CookieHttpOnly = s.App.GetBoolConfig("CookieHttpOnly", false)
		CookieSecure = s.App.GetBoolConfig("CookieSecure", false)
		SessionTimeout = int(s.App.GetIntConfig("SessionTimeout", 30))
	})
}

type MemSessionManager struct {
	peony.SessionManager
	sessions map[string]*item
	mtx      sync.Mutex
	list     *list.List
}

func (sm *MemSessionManager) GenerateId() string {
	uuid, err := simpleuuid.NewTime(time.Now())
	if err != nil {
		panic(err) // never happend
	}
	return uuid.String()
}

func (sm *MemSessionManager) clearTask() {
	for {
		select {
		case <-time.After(60 * time.Second):
		}
		now := time.Now().Unix()
		sm.mtx.Lock()
		for e := sm.list.Front(); e != nil; {
			wrap := e.Value.(*sessionWrap)
			if (now - wrap.lastAccess) >= SessionTimeout*60 {
				rm := e
				e = e.Next()
				sm.list.Remove(rm)
				delete(sm.sessions, wrap.GetId())
				continue
			} else {
				break
			}
			e = e.Next()
		}
		sm.mtx.Unlock()
	}
}

func (sm *MemSessionManager) Store(c *peony.Controller, s *peony.Session) {
	sm.saveSession(s)
	cookie := &http.Cookie{
		Name:     "PEONY_SESSIONID",
		HttpOnly: CookieSecure,
		Secure:   CookieHttpOnly,
		Value:    s.GetId(),
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(SessionTimeout) * time.Second).UTC(),
	}
	http.SetCookie(c.Resp, cookie)
}

func (sm *MemSessionManager) saveSession(session *peony.Session)  {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()
	it := sm.sessions[session.GetId()]
	if it == nil {
		it = &item{}
		wrap := &sessionWrap{Session: session}
		el := sm.list.PushBack(wrap)
		it.elem = el
		it.wrap = wrap
		sm.sessions[session.GetId()] = it
	} else {
		sm.list.MoveToBack(it.elem)
	}
	it.wrap.lastAccess = time.Now().Unix()
}

func (sm *MemSessionManager) getSession(id string) *peony.Session {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()
	var session *peony.Session
	if len(id) != 0 {
		it := sm.sessions[id]
		if it != nil {
			session = it.wrap.Session	
		}
	}
	if session == nil {
		session = &peony.Session{
			Attribute: make(map[string]interface{}), 
			Id: sm.GenerateId(),
		}
	}
	return session
}


func (sm *MemSessionManager) Get(c *peony.Controller) *peony.Session {
	cookie, err := c.Req.Cookie("PEONY_SESSIONID")
	var session *peony.Session
	var id string
	if err == nil {
		id = cookie.Value
	}
	session = sm.getSession(id)
	return session
}
