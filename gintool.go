package gintool

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// GinEngine is the configuration of gin.Engine
type GinEngine struct {
	Engine   *gin.Engine
	address  string
	statics  map[string]string
	staticFs map[string]string
	errors   map[int]string
	logfile  string
	errorlog string
	certFile string
	keyFile  string
	server   *http.Server
	template multitemplate.Renderer
}

var stdlog zerolog.Logger = zlog.Output(os.Stdout)
var errlog zerolog.Logger = zlog.Output(os.Stderr)

func fileInfo(path string) os.FileInfo {
	fi, err := os.Stat(path)
	if err != nil {
		return nil
	}
	return fi
}

func isFile(path string) error {
	fi := fileInfo(path)
	if fi == nil {
		return fmt.Errorf("not found %v", path)
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("not a file %v", path)
	}
	return nil
}

// NewGin will create a new GinEngine with the config file.
// Example config files check the config and testdata directory.
func NewGin(path string) (*GinEngine, error) {
	resetDefault()
	if path == "" {
		path = "config/gin.conf"
	}
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	gin.SetMode(gin.DebugMode)

	ge := &GinEngine{
		Engine:  engine,
		address: "localhost:8080",
		errors:  make(map[int]string),
		statics: map[string]string{},
	}
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
	//fmt.Println("unmarshal: ", out)
	m, err := extract(out, "gin")
	if err != nil {
		return nil, err
	}
	mm, err := extract(m, "address")
	if err == nil {
		ge.address = mm.(string)
		//fmt.Println("address:", mm)
	}
	mm, err = extract(m, "log")
	if err == nil {
		ge.logfile = mm.(string)
		//fmt.Println("logfile:", mm)
	}
	mm, err = extract(m, "errorlog")
	if err == nil {
		ge.errorlog = mm.(string)
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
			ge.certFile = ss
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
			ge.keyFile = ss
		}
	}
	if ge.certFile == "" || ge.keyFile == "" {
		ge.certFile = ""
		ge.keyFile = ""
		//} else {
		//fmt.Println("certfile", ge.certFile)
		//fmt.Println("keyfile", ge.keyFile)
	}
	mm, err = extract(m, "static")
	if err == nil {
		//ge.statics = mm.(string)
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
		ge.statics = static
		//fmt.Println("statics", ge.statics)
	}
	mm, err = extract(m, "staticfile")
	if err == nil {
		ss, ok := mm.([]interface{})
		if !ok {
			return nil, fmt.Errorf("wrong type staticfile")
		}
		static := make(map[string]string)
		for _, s := range ss {
			//fmt.Println("statics", s)
			mapping, ok := s.(map[interface{}]interface{})
			if !ok {
				return nil, fmt.Errorf("wrong type staticfile mapping %v", s)
			}
			static[mapping["map"].(string)] = mapping["file"].(string)
		}
		ge.staticFs = static
		//fmt.Println("staticfile", ge.staticFs)
	}
	mm, err = extract(m, "error", "404")
	//fmt.Println("err", err)
	if err == nil {
		ge.errors[http.StatusNotFound] = mm.(string)
		//fmt.Println("404:", mm)
	}
	mm, err = extract(m, "error", "500")
	if err == nil {
		ge.errors[http.StatusInternalServerError] = mm.(string)
		//fmt.Println("500:", mm)
	}
	ge.template = multitemplate.NewRenderer()
	gin.ForceConsoleColor()
	var logs io.Writer
	var logfile *os.File
	if len(ge.logfile) > 0 {
		logfile, err = os.OpenFile(ge.logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err == nil {
			if gin.Mode() != gin.ReleaseMode {
				logs = io.MultiWriter(logfile, os.Stdout)
			} else {
				logs = io.MultiWriter(logfile)
			}
		}
	}

	if gin.IsDebugging() {
		stdlog.Level(zerolog.DebugLevel)
	}

	if logs != nil {
		stdlog = stdlog.Output(
			zerolog.ConsoleWriter{
				Out:     logs,
				NoColor: false,
			},
		)
	}

	gin.DefaultWriter = stdlog

	logs = errlog
	if len(ge.errorlog) > 0 {
		if ge.errorlog != ge.logfile {
			logfile, err = os.OpenFile(ge.errorlog, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
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
		errlog = errlog.Output(
			zerolog.ConsoleWriter{
				Out:     logs,
				NoColor: false,
			},
		)
	}
	gin.DefaultErrorWriter = errlog

	engine.Use(ginRecovery(ge.errors))
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		stdlog.Level(zerolog.DebugLevel)
		errlog.Level(zerolog.DebugLevel)
	}

	engine.Use(logger.SetLogger(logger.Config{
		Logger: &stdlog,
		UTC:    true,
	}))
	engine.Use(UseSession())
	return ge, nil
}

func (ge *GinEngine) AddTemplates(name string, files ...string) {
	ge.template.AddFromFiles(name, files...)
}

func resetDefault() {
	gin.DefaultWriter = os.Stdout
	gin.DefaultErrorWriter = os.Stderr
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	stdlog = zlog.Output(os.Stdout)
	errlog = zlog.Output(os.Stderr)
}

// ShutDown will shutdown the engine
func (g *GinEngine) ShutDown() (err error) {
	if g.server == nil {
		return fmt.Errorf("server not start")
	}
	defer func() {
		if err != nil {
			return
		}
		stdlog.Info().Msgf("**********************")
		stdlog.Info().Msgf("* shutdown %s *", g.address)
		stdlog.Info().Msgf("**********************\n")
	}()
	return g.server.Shutdown(context.Background())
}

// Start just start the engine, tls will according to the configuration file
func (g *GinEngine) Start() (ret error) {
	for mapping, path := range g.statics {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		//fi := fileInfo(path)
		//fmt.Println(mapping, path, " ==== ", fi)
		g.Engine.Static(mapping, path)
	}
	for mapping, path := range g.staticFs {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		//fi := fileInfo(path)
		//fmt.Println(mapping, path, " ==== ", fi)
		g.Engine.StaticFile(mapping, path)
	}

	// TODO add template path

	// add error handling
	for key, value := range g.errors {
		g.AddTemplates(errorName(key), value)
	}
	stdlog.Debug().Msgf("errors : %s %v", g.errors[http.StatusNotFound], g.errors)
	if _, ok := g.errors[http.StatusNotFound]; ok {
		g.Engine.NoRoute(func(c *gin.Context) {
			c.HTML(http.StatusNotFound, errorName(http.StatusNotFound), nil)
		})
	}
	g.Engine.HTMLRender = g.template

	stdlog.Info().Msgf("| starting gin server |")
	stdlog.Info().Msgf("=======================")
	stdlog.Info().Msgf("| tls     : %v", g.certFile != "")
	stdlog.Info().Msgf("| mode    : %s", gin.Mode())
	stdlog.Info().Msgf("| address : %s", g.address)
	if g.logfile != "" {
		stdlog.Info().Msgf("| logfile : %s", g.logfile)
	}
	if g.errorlog != "" {
		stdlog.Info().Msgf("| errorlog: %s", g.errorlog)
	}
	if len(g.statics) > 0 {
		stdlog.Info().Msgf("| statics : %v", g.statics)
	}
	if len(g.staticFs) > 0 {
		stdlog.Info().Msgf("| staticFs: %v", g.staticFs)
	}
	if len(g.errors) > 0 {
		stdlog.Info().Msgf("| errors  : %v", g.errors)
	}
	stdlog.Info().Msgf("=======================")
	//defer func() { debugPrintError(err) }()

	address := g.address
	runtype := "HTTP"
	if g.certFile != "" {
		runtype = "HTTPS"
	}
	stdlog.Info().Msgf("Listening and serving %s on %s\n", runtype, address)
	server := http.Server{Addr: address, Handler: g.Engine}

	defer func() {
		g.server = nil
	}()
	g.server = &server
	if g.certFile != "" {
		ret = server.ListenAndServeTLS(g.certFile, g.keyFile)
	} else {
		ret = server.ListenAndServe()
	}
	return
}

func ginRecovery(errors map[int]string) gin.HandlerFunc {
	return RecoveryWithWriter(func(c *gin.Context, err interface{}) {
		if _, ok := errors[c.Writer.Status()]; ok {
			c.HTML(c.Writer.Status(), errorName(c.Writer.Status()), gin.H{
				"errors": errors,
			})
		}
	})
}

func RecoveryWithWriter(f func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := errors.Wrap(err, 0)
				msg := "[" + stack.Error() + "]"

				status := http.StatusInternalServerError
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
							msg = "[broken pipe]"
							status = http.StatusGatewayTimeout
						}
					}
				}
				// save log to stdlog
				dumplogger := stdlog.With().
					Int("status", status).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("ip", c.ClientIP()).
					Dur("latency", 0).
					Str("user-agent", c.Request.UserAgent()).
					Logger()

				dumplogger.Error().Msg(msg)

				c.AbortWithStatus(status)

				if (brokenPipe) {
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
				errlog.Info().Msgf("errors when visit: %s", c.Request.URL.Path)
				if gin.IsDebugging() {
					errlog.Error().Msgf("[Recovery] panic recovered:\n%s", h)
					errlog.Error().Msgf("[Recovery] [%s]\n%s", err, stack.Stack())
				} else {
					errlog.Error().Msgf("[Recovery] panic recovered:\n[%s] %s\n%s",
						c.Request.URL.Path, err, stack.Stack())
				}

				f(c, err)
			}
		}()
		c.Next() // execute all the handlers
	}
}

// HandleSession will create a new session with key map to store the value for future use.
// For example, you can store the language define in session then use it in template or i18n.
// Warning: you should not use session in middleware because it will be called after the middleware
func (g *GinEngine) HandleSession(method string, path string, handlerFunc gin.HandlerFunc) gin.IRoutes {
	return g.Engine.Handle(method, path, func(c *gin.Context) {
		WithSession(func() {
			handlerFunc(c)
		})()
	})
}
