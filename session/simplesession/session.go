package simple_session

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"github.com/joinhack/peony"
	"github.com/streadway/simpleuuid"
	"net/http"
	"time"
)

var (
	encoding *base64.Encoding
	CookieHttpOnly bool
	CookieSecure bool
	SessionTimeout int
)

func init() {
	gob.Register((*peony.Session)(nil))

	peony.OnServerInit(func(s *peony.Server) {
		sec := s.App.Security
		encoding = base64.NewEncoding(sec)
		s.RegisterSessionManager(&SimpleSessionManager{})
		CookieHttpOnly = s.App.GetBoolConfig("CookieHttpOnly", false)
		CookieSecure = s.App.GetBoolConfig("CookieSecure", false)
		SessionTimeout = s.App.GetIntConfig("SessionTimeout", 30)
	})
}

type SimpleSessionManager struct {
	peony.SessionManager
}

func (c *SimpleSessionManager) GenerateId() string {
	uuid, err := simpleuuid.NewTime(time.Now())
	if err != nil {
		panic(err) // never happend
	}
	return uuid.String()
}

func (cm *SimpleSessionManager) Store(c *peony.Controller, s *peony.Session) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		peony.ERROR.Println(err)
	}

	val := encoding.EncodeToString(buf.Bytes())
	cookie := &http.Cookie{
		Name:     "PEONY_SESSION",
		HttpOnly: CookieHttpOnly,
		Secure:   CookieSecure,
		Value:    val,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(SessionTimeout) * time.Second).UTC(),
	}
	http.SetCookie(c.Resp, cookie)
}

func (sm *SimpleSessionManager) Get(c *peony.Controller) *peony.Session {
	cookie, err := c.Req.Cookie("PEONY_SESSION")
	if err != nil {
		return &peony.Session{Attribute: make(map[string]interface{}), Id: sm.GenerateId()}
	}
	var bs []byte
	bs, err = encoding.DecodeString(cookie.Value)
	if err != nil {
		peony.ERROR.Println(err)
		return &peony.Session{Attribute: make(map[string]interface{}), Id: sm.GenerateId()}
	}
	var buf = bytes.NewBuffer(bs)
	enc := gob.NewDecoder(buf)
	var session = &peony.Session{}
	if err = enc.Decode(session); err != nil {
		peony.ERROR.Println(err)
		return &peony.Session{Attribute: make(map[string]interface{}), Id: sm.GenerateId()}
	}
	return session
}
