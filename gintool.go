// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gintool

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"

	"github.com/FlowerWrong/plushgin"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog"
	//zlog "github.com/rs/zerolog/log"
)

// GinEngine is the configuration of gin.Engine
type GinEngine struct {
	// Engine is the exposed *gin.Engine
	Engine   *gin.Engine
	server   *http.Server
	template *plushgin.Plush2Render
	config   *Config
}

//var stdlog = zlog.Output(os.Stdout)
//var errlog = zlog.Output(os.Stderr)

// NewGin will create a new GinEngine with the config file.
// Example config files check the config and testdata directory.
func NewGin(path string) (*GinEngine, error) {
	resetDefault()
	if path == "" {
		path = "config/gin.conf"
	}
	c, e := parseFile(path)
	if e != nil {
		return nil, e
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	gin.SetMode(gin.DebugMode)

	ge := &GinEngine{
		Engine: engine,
		config: c,
	}

	ge.template = plushgin.Default()
	gin.ForceConsoleColor()
	var logs io.Writer
	var logfile *os.File
	if len(c.logfile) > 0 {
		logfile, err := os.OpenFile(c.logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err == nil {
			if gin.Mode() != gin.ReleaseMode {
				logs = io.MultiWriter(logfile, os.Stdout)
			} else {
				logs = io.MultiWriter(logfile)
			}
		}
	}

	if gin.IsDebugging() {
		c.stdlog.Level(zerolog.DebugLevel)
	}

	if logs == nil {
		logs = os.Stdout
	}
	c.stdlog = c.stdlog.Output(
		zerolog.ConsoleWriter{
			Out:     logs,
			NoColor: false,
		},
	).With().Caller().CallerWithSkipFrameCount(2).Logger()

	gin.DefaultWriter = c.stdlog

	logs = nil
	if len(c.errorlog) > 0 {
		if c.errorlog != c.logfile {
			logfile, err := os.OpenFile(c.errorlog, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err == nil {
				if gin.Mode() != gin.ReleaseMode {
					logs = io.MultiWriter(logfile, os.Stderr)
				} else {
					logs = io.MultiWriter(logfile)
				}
			}
		} else {
			if gin.Mode() != gin.ReleaseMode {
				logs = io.MultiWriter(logfile, os.Stderr)
			} else {
				logs = io.MultiWriter(logfile)
			}
		}
	}
	if logs == nil {
		logs = os.Stderr
	}
	c.errlog = c.errlog.Output(
		zerolog.ConsoleWriter{
			Out:     logs,
			NoColor: false,
		},
	).With().Caller().CallerWithSkipFrameCount(2).Logger()
	gin.DefaultErrorWriter = c.errlog

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		c.stdlog.Level(zerolog.DebugLevel)
		c.errlog.Level(zerolog.DebugLevel)
	}

	engine.Use(logger.SetLogger(logger.Config{
		Logger: &c.stdlog,
		UTC:    true,
	}))
	engine.Use(ginRecovery(c.errors, c))
	engine.Use(UseSession(c))
	return ge, nil
}

// AddTemplates to add templates with the specified name
func (ge *GinEngine) AddTemplates(name string, files ...string) {
	//ge.template.Options..AddFromFiles(name, files...)
}

func resetDefault() {
	gin.DefaultWriter = os.Stdout
	gin.DefaultErrorWriter = os.Stderr
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	//stdlog = zlog.Output(os.Stdout)
	//errlog = zlog.Output(os.Stderr)
}

// ShutDown will shutdown the engine
func (ge *GinEngine) ShutDown() (err error) {
	if ge.server == nil {
		return fmt.Errorf("server not start")
	}
	defer func() {
		if err != nil {
			return
		}
		ge.config.stdlog.Info().Msgf("**********************")
		ge.config.stdlog.Info().Msgf("* shutdown %s *", ge.config.address)
		ge.config.stdlog.Info().Msgf("**********************\n")
	}()
	return ge.server.Shutdown(context.Background())
}

// Start just start the engine, tls will according to the configuration file
func (ge *GinEngine) Start() (ret error) {
	c := ge.config
	if c == nil {
		c = initConfig()
		ge.config = c
	}
	for mapping, path := range c.statics {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		//fi := fileInfo(path)
		//fmt.Println(mapping, path, " ==== ", fi)
		ge.Engine.Static(mapping, path)
	}
	for mapping, path := range c.staticFs {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		//fi := fileInfo(path)
		//fmt.Println(mapping, path, " ==== ", fi)
		ge.Engine.StaticFile(mapping, path)
	}

	if c.templates != "" {
		ge.template.Options.TemplateDir = c.templates
	}

	// add error handling
	//for key, value := range c.errors {
	//	ge.AddTemplates(errorName(key), value)
	//}
	c.stdlog.Debug().Msgf("errors : %s %v", ge.config.errors[http.StatusNotFound], c.errors)
	if _, ok := ge.config.errors[http.StatusNotFound]; ok {
		ge.Engine.NoRoute(func(c *gin.Context) {
			c.HTML(http.StatusNotFound, ge.config.errors[http.StatusNotFound], gin.H{})
		})
	}
	ge.Engine.NoMethod()
	ge.Engine.HTMLRender = ge.template

	c.stdlog.Info().Msgf("| starting gin server |")
	c.stdlog.Info().Msgf("=======================")
	c.stdlog.Info().Msgf("| tls     : %v", c.certFile != "")
	c.stdlog.Info().Msgf("| mode    : %s", gin.Mode())
	c.stdlog.Info().Msgf("| address : %s", c.address)
	if c.logfile != "" {
		c.stdlog.Info().Msgf("| logfile : %s", c.logfile)
	}
	if c.errorlog != "" {
		c.stdlog.Info().Msgf("| errorlog: %s", c.errorlog)
	}
	if c.templates != "" {
		c.stdlog.Info().Msgf("| templates: %s", c.templates)
	}
	if len(c.statics) > 0 {
		c.stdlog.Info().Msgf("| statics : %v", c.statics)
	}
	if len(c.staticFs) > 0 {
		c.stdlog.Info().Msgf("| staticFs: %v", c.staticFs)
	}
	if len(c.errors) > 0 {
		c.stdlog.Info().Msgf("| errors  : %v", c.errors)
	}
	c.stdlog.Info().Msgf("=======================")
	//defer func() { debugPrintError(err) }()

	address := c.address
	runtype := "HTTP"
	if c.certFile != "" {
		runtype = "HTTPS"
	}
	c.stdlog.Info().Msgf("Listening and serving %s on %s\n", runtype, address)
	server := http.Server{Addr: address, Handler: ge.Engine}

	defer func() {
		ge.server = nil
	}()
	ge.server = &server
	if c.certFile != "" {
		ret = server.ListenAndServeTLS(c.certFile, c.keyFile)
	} else {
		ret = server.ListenAndServe()
	}
	if ret == http.ErrServerClosed {
		ret = nil
	}
	return
}

func ginRecovery(errors map[int]string, cc *Config) gin.HandlerFunc {
	return recoveryWithWriter(cc, func(c *gin.Context) {
		if v, ok := errors[c.Writer.Status()]; ok {
			c.HTML(c.Writer.Status(), v, gin.H{
				"errors": errors,
			})
		}
	})
}

func recoveryWithWriter(config *Config, f func(c *gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := errors.Wrap(err, 2)
				status := http.StatusInternalServerError
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
							err = errors.New("[broken pipe]")
							status = http.StatusGatewayTimeout
						}
					}
				}

				_ = c.AbortWithError(status, stack)

				if brokenPipe {
					return
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}
				h := strings.Join(headers, "\n")
				config.errlog.Info().Msgf("errors when visit: %s", c.Request.URL.Path)
				if gin.IsDebugging() {
					config.errlog.Error().Msgf("[Recovery] panic recovered:\n%s", h)
					config.errlog.Error().Msgf("[Recovery] [%s]\n%s", err, stack.Stack())
				} else {
					config.errlog.Error().Msgf("[Recovery] panic recovered:\n[%s] %s\n%s",
						c.Request.URL.Path, err, stack.Stack())
				}

				f(c)
			}
		}()
		c.Next() // execute all the handlers
	}
}

// HandleSession will create a new session with key map to store the value for future use.
// For example, you can store the language define in session then use it in template or i18n.
// Warning: you should not use session in middleware because it will be called after the middleware
func (ge *GinEngine) HandleSession(method string, path string, handlerFunc gin.HandlerFunc) gin.IRoutes {
	return ge.Engine.Handle(method, path, func(c *gin.Context) {
		WithSession(func() {
			handlerFunc(c)
		})()
	})
}
