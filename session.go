package peony

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"github.com/streadway/simpleuuid"
	"net/http"
	"time"
)

var (
	encoding *base64.Encoding
)

type Session struct {
	Attribute map[string]interface{}
	Id        string
}

//Set attribute
func (s *Session) Set(key string, value interface{}) {
	s.Attribute[key] = value
}

//Get attribute
func (s *Session) Get(key string) (val interface{}, ok bool) {
	val, ok = s.Attribute[key]
	return
}

//Get session id
func (s *Session) GetId() string {
	return s.Id
}

//Set session id
func (s *Session) SetId(id string) {
	s.Id = id
}

type SessionManager interface {
	GenerateId() string
	Store(c *Controller, s *Session) //save session
	Get(c *Controller) *Session      //get session
}

type SimpleSessionManager struct {
	SessionManager
}

func (c *SimpleSessionManager) GenerateId() string {
	uuid, err := simpleuuid.NewTime(time.Now())
	if err != nil {
		panic(err) // never happend
	}
	return uuid.String()
}

func (cm *SimpleSessionManager) Store(c *Controller, s *Session) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		ERROR.Println(err)
	}

	val := encoding.EncodeToString(buf.Bytes())
	cookie := &http.Cookie{
		Name:     "PEONY_SESSION",
		HttpOnly: false,
		Secure:   false,
		Value:    val,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Second).UTC(),
	}
	http.SetCookie(c.Resp, cookie)
}

func (cm *SimpleSessionManager) Get(c *Controller) *Session {
	cookie, err := c.Req.Cookie("PEONY_SESSION")
	if err != nil {
		return &Session{Attribute: make(map[string]interface{}), Id: cm.GenerateId()}
	}
	var bs []byte
	bs, err = encoding.DecodeString(cookie.Value)
	if err != nil {
		ERROR.Println(err)
		return &Session{Attribute: make(map[string]interface{}), Id: cm.GenerateId()}
	}
	var buf = bytes.NewBuffer(bs)
	enc := gob.NewDecoder(buf)
	var session = &Session{}
	if err = enc.Decode(session); err != nil {
		ERROR.Println(err)
		return &Session{Attribute: make(map[string]interface{}), Id: cm.GenerateId()}
	}
	return session
}

//default sessionManager use the cookie session manager
func init() {
	OnServerInit(func(s *Server) {
		s.SessionManager = &SimpleSessionManager{}
		gob.Register((*Session)(nil))
		sec := s.App.Security
		if len(sec) == 0 {
			sec = defaultSecKey
		}
		encoding = base64.NewEncoding(sec)
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
