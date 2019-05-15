// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gintool

import (
	"github.com/gin-gonic/gin"
	"github.com/v2pro/plz/gls"
)

const session_name string = "_session_"
const config_name string = "_config_"

type Session struct {
	data map[string]interface{}
}

// GetSession get the Session, if not exist return new
func GetSession() *Session {
	if v := gls.Get(session_name); v != nil {
		return v.(*Session)
	}
	return nil
}

// NewSession create new Session
func NewSession() *Session {
	gls.GoID()
	var s *Session
	if s = GetSession(); s == nil {
		s = &Session{
			data: map[string]interface{}{},
		}
		gls.Set(session_name, s)
	}
	return s
}

// SessionGet return the value stored in Session
func SessionGet(name string) interface{} {
	s := GetSession()
	return s.data[name]
}

// SessionSet store the value to Session
func SessionSet(name string, val interface{}) {
	s := GetSession()
	s.data[name] = val
	//glsSetSession(s)
}

// UseSession is a middleware which generate the session in gorouting
// Please be sure that all middleware use session must called after this middleware
// GinEngine default will use this middleware
func UseSession(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		WithSession(func() {
			config.stdlog.Debug().Msgf("session initialed %v %v", gls.GoID(), config)
			SessionSet(config_name, &config)
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

// SessionConfig return the saved *Config, it must be called after Use(UseSession(*Config))
func SessionConfig() *Config {
	t := SessionGet(config_name)
	tt, ok := t.(**Config)
	if ok {
		return *tt
	}
	return nil
}
