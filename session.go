package gintool

import (
	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
	"github.com/v2pro/plz/gls"
)

const SESSION_NAME string = "_session_"
type Session struct {
	data map[string]interface{}
}

func GetSession() *Session {
	if v := gls.Get(SESSION_NAME); v != nil {
		return v.(*Session)
	}
	s := NewSession()
	return s
}

func NewSession() *Session {
	s := &Session{
		data: map[string]interface{}{},
	}
	gls.GoID()
	gls.Set(SESSION_NAME, s)
	return s
}

func SessionGet(name string) interface{} {
	s := GetSession()
	return s.data[name]
}

func SessionSet(name string, val interface{}) {
	s := GetSession()
	s.data[name] = val
}

// UseSession is a middleware which generate the session in gorouting
// Please be sure that all middleware use session must called after this middleware
// GinEngine default will use this middleware
func UseSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		WithSession(func() {
			zlog.Print("session initialed ", gls.GoID())
			c.Next()
		})()
	}
}

// WithSession will create the session key in current gorouting
func WithSession(f func()) func() {
	return gls.WithGls(func() {
		NewSession()
		f()
	})
}
