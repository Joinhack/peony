package peony

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

var (
	cookieKeyValueParser = regexp.MustCompile("\x00([^:]*):([^\x00]*)\x00")
	CookiePrefix         = "PEONY"
	CookieHttpOnly       = false
	CookieSecure         = false
)

func ParseKeyValueCookie(val string, cb func(key, val string)) {
	val, _ = url.QueryUnescape(val)
	if matches := cookieKeyValueParser.FindAllStringSubmatch(val, -1); matches != nil {
		for _, match := range matches {
			cb(match[1], match[2])
		}
	}
}

type Flash struct {
	In, Out map[string]string
}

func (f Flash) Error(msg string, args ...interface{}) {
	if len(args) == 0 {
		f.Out["error"] = msg
	} else {
		f.Out["error"] = fmt.Sprintf(msg, args...)
	}
}

// Success serializes the given msg and args to a "success" key within
// the Flash cookie.
func (f Flash) Success(msg string, args ...interface{}) {
	if len(args) == 0 {
		f.Out["success"] = msg
	} else {
		f.Out["success"] = fmt.Sprintf(msg, args...)
	}
}

func GetFlashFilter(s *Server) Filter {
	CookieHttpOnly = s.App.GetBoolConfig("CookieHttpOnly", false)
	CookieSecure = s.App.GetBoolConfig("CookieSecure", false)
	CookiePrefix = s.App.GetStringConfig("CookiePrefix", "PEONY")
	return func (c *Controller, fc []Filter) {
		c.Flash = restoreFlash(c.Req.Request)

		fc[0](c, fc[1:])

		// Store the flash.
		var flashValue string
		for key, value := range c.Flash.Out {
			flashValue += "\x00" + key + ":" + value + "\x00"
		}
		c.SetCookie(&http.Cookie{
			Name:     CookiePrefix + "_FLASH",
			Value:    url.QueryEscape(flashValue),
			HttpOnly: CookieHttpOnly,
			Secure:   CookieSecure,
			Path:     "/",
		})
	}
}

func restoreFlash(req *http.Request) *Flash {
	flash := &Flash{
		In: make(map[string]string),
		Out:  make(map[string]string),
	}
	if cookie, err := req.Cookie(CookiePrefix + "_FLASH"); err == nil {
		ParseKeyValueCookie(cookie.Value, func(key, val string) {
			flash.In[key] = val
		})
	}
	return flash
}
