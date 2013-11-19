package peony

import (
	"github.com/streadway/simpleuuid"
	"net/http"
	"time"
)

type Session struct {
	attribute map[string]interface{}
	id        string
}

func (s *Session) Set(key string, value interface{}) {
	s.attribute[key] = value
}

func (s *Session) Get(key string) (val interface{}, ok bool) {
	val, ok = s.attribute[key]
	return
}

func (s *Session) Id() string {
	return s.id
}

func (s *Session) SetId(id string) {
	s.id = id
}

type SessionManager interface {
	GenerateId() string
	Save(c *Controller, s *Session)
	Get(c *Controller) *Session
}

type CookieSessionManager struct {
	SessionManager
}

func (c *CookieSessionManager) GenerateId() string {
	uuid, err := simpleuuid.NewTime(time.Now())
	if err != nil {
		panic(err) // never happend
	}
	return uuid.String()
}

func (cm *CookieSessionManager) Save(c *Controller, s *Session) {
	cookie := &http.Cookie{
		Name:     "PEONY_SID",
		HttpOnly: false,
		Secure:   false,
		Value:    "asdasds",
		Path:     "/",
		Expires:  time.Now().UTC(),
	}
	http.SetCookie(c.Resp, cookie)
}

func (cm *CookieSessionManager) Get(c *Controller) *Session {
	cookie, err := c.Req.Cookie("PEONY_SID")
	if err != nil {
		return &Session{attribute: make(map[string]interface{}), id: cm.GenerateId()}
	}
	println(cookie)
	return nil
}

//default sessionManager use the cookie session manager
func init() {
	OnServerInit(func(s *Server) {
		s.SetSessionManager(&CookieSessionManager{})
	})
}

func GetSessionFilter(sm SessionManager) Filter {
	return func(c *Controller, filter []Filter) {
		session := sm.Get(c)
		c.Session = session
		filter[0](c, filter[1:])
		sm.Save(c, session)
	}
}
