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

//Set attribute
func (s *Session) Set(key string, value interface{}) {
	s.attribute[key] = value
}

//Get attribute
func (s *Session) Get(key string) (val interface{}, ok bool) {
	val, ok = s.attribute[key]
	return
}

//Get session id
func (s *Session) Id() string {
	return s.id
}

//Set session id
func (s *Session) SetId(id string) {
	s.id = id
}

type SessionManager interface {
	GenerateId() string
	Store(c *Controller, s *Session) //save session
	Get(c *Controller) *Session      //get session
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
		s.SessionManager = &CookieSessionManager{}
	})
}

func GetSessionFilter(s *Server) Filter {
	return func(c *Controller, filter []Filter) {
		session := s.SessionManager.Get(c)
		c.Session = session
		filter[0](c, filter[1:])
		s.SessionManager.Store(c, session)
	}
}
