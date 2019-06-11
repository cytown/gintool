// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gintool

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config the configuration
type Config struct {
	address   string
	mode      string
	statics   map[string]string
	staticFs  map[string]string
	errors    map[int]string
	templates string
	logfile   string
	errorlog  string
	certFile  string
	keyFile   string
	stdlog    zerolog.Logger
	errlog    zerolog.Logger
	other     interface{}
}

func initConfig() *Config {
	return &Config{
		errors:   map[int]string{},
		statics:  map[string]string{},
		staticFs: map[string]string{},
		stdlog:   log.Output(os.Stdout),
		errlog:   log.Output(os.Stderr),
	}
}

func parseFile(path string) (*Config, error) {
	err := isFile(path)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	//fmt.Println(len(buf))
	var out interface{}
	err = yaml.Unmarshal(buf, &out)
	if err != nil {
		return nil, err
	}
	c := initConfig()
	//fmt.Println("unmarshal: ", out)
	m, err := extract(out, "gin")
	if err != nil {
		return nil, err
	}
	mm, err := extract(m, "address")
	if err == nil {
		c.address = mm.(string)
		//fmt.Println("address:", mm)
	}
	mm, err = extract(m, "other")
	if err == nil {
		c.other = mm
		//fmt.Println("address:", mm)
	}
	mm, err = extract(m, "log")
	if err == nil {
		c.logfile = mm.(string)
		//fmt.Println("logfile:", mm)
	}
	mm, err = extract(m, "errorlog")
	if err == nil {
		c.errorlog = mm.(string)
		//fmt.Println("errorlog:", mm)
	}
	mm, err = extract(m, "mode")
	if err == nil {
		gin.SetMode(mm.(string))
		//fmt.Println("mode:", mm)
	}
	mm, err = extract(m, "tls", "certfile")
	if err == nil {
		ss, ok := mm.(string)
		if ok && ss != "" {
			err := isFile(ss)
			if err != nil {
				return nil, err
			}
			c.certFile = ss
		}
	}
	mm, err = extract(m, "tls", "keyfile")
	if err == nil {
		ss, ok := mm.(string)
		if ok && ss != "" {
			err := isFile(ss)
			if err != nil {
				return nil, err
			}
			c.keyFile = ss
		}
	}
	if c.certFile == "" || c.keyFile == "" {
		c.certFile = ""
		c.keyFile = ""
		//} else {
		//fmt.Println("certfile", c.certFile)
		//fmt.Println("keyfile", c.keyFile)
	}
	mm, err = extract(m, "templates")
	if err == nil {
		ss, ok := mm.(string)
		if ok && ss != "" {
			err := isDir(ss)
			if err != nil {
				return nil, err
			}
			c.templates = ss
		}
	}
	mm, err = extract(m, "static")
	if err == nil {
		//c.statics = mm.(string)
		//fmt.Println("statics:", mm)
		ss, ok := mm.([]interface{})
		if !ok {
			return nil, fmt.Errorf("wrong type statics")
		}
		static := make(map[string]string)
		for _, s := range ss {
			//fmt.Println("statics", s)
			mapping, ok := s.(map[interface{}]interface{})
			if !ok {
				return nil, fmt.Errorf("wrong type statics mapping %v", s)
			}
			static[mapping["map"].(string)] = mapping["path"].(string)
		}
		c.statics = static
		//fmt.Println("statics", c.statics)
	}
	mm, err = extract(m, "staticfile")
	if err == nil {
		ss, ok := mm.([]interface{})
		if !ok {
			return nil, fmt.Errorf("wrong type staticfile")
		}
		static := map[string]string{}
		for _, s := range ss {
			//fmt.Println("statics", s)
			mapping, ok := s.(map[interface{}]interface{})
			if !ok {
				return nil, fmt.Errorf("wrong type staticfile mapping %v", s)
			}
			static[mapping["map"].(string)] = mapping["file"].(string)
		}
		c.staticFs = static
		//fmt.Println("staticfile", c.staticFs)
	}
	mm, err = extract(m, "error", "404")
	//fmt.Println("err", err)
	if err == nil {
		c.errors[http.StatusNotFound] = mm.(string)
		//fmt.Println("404:", mm)
	}
	mm, err = extract(m, "error", "500")
	if err == nil {
		c.errors[http.StatusInternalServerError] = mm.(string)
		//fmt.Println("500:", mm)
	}
	return c, nil
}

// Get return the saved other configuration with key/value mapping
func (c *Config) Get(key ...string) (ret interface{}) {
	ret, _ = extract(c.other, key...)
	return
}
