// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
// The original package is from "github.com/FlowerWrong/plushgin" please respect him

package plushgin

import (
	"io/ioutil"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/gobuffalo/plush"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const htmlContentType = "text/html; charset=utf-8"

// RenderOptions is used to configure the renderer.
type RenderOptions struct {
	TemplateDir     string
	ContentType     string
	MaxCacheEntries int
}

// Plush2Render is a custom Gin template renderer using plush.
type Plush2Render struct {
	Options *RenderOptions
	Name    string
	Context plush.Context
	cache   *templateCache
	helpers map[string]interface{}
}

// New creates a new Plush2Render instance with custom Options.
func New(options RenderOptions) *Plush2Render {
	p := Plush2Render{
		Options: &options,
		cache:   newTemplateCache(options.MaxCacheEntries),
	}
	log.Logger.Level(zerolog.DebugLevel)
	p.initDefaultHelpers()
	return &p
}

// Default creates a Plush2Render instance with default options.
func Default() *Plush2Render {
	return New(RenderOptions{
		TemplateDir:     "templates",
		ContentType:     htmlContentType,
		MaxCacheEntries: 128,
	})
}

// Instance should return a new Plush2Render struct per request and prepare
// the template by either loading it from disk or using plush's cache.
func (p *Plush2Render) Instance(name string, data interface{}) render.Render {
	log.Logger.Level(zerolog.DebugLevel)
	return &Plush2Render{
		Context: NewContext(p, data.(gin.H)),
		Options: p.Options,
		cache:   p.cache,
		Name:    name,
	}
}

// Render should render the template to the response.
func (p *Plush2Render) Render(w http.ResponseWriter) error {
	var err error
	var renderedStr string

	buf, err := p.getCache(p.Name)
	if err != nil {
		panic(err)
	}
	renderedStr, err = plush.Render(string(buf), &p.Context)
	if err != nil {
		panic(err)
	}
	rendered := []byte(renderedStr)
	p.WriteContentType(w)
	_, err = w.Write(rendered)
	return err
}

func (p *Plush2Render) getCache(name string) ([]byte, error) {
	buf := p.cache.Get(name)
	if buf == nil || gin.Mode() == "debug" {
		filename := path.Join(p.Options.TemplateDir, name)
		var err error
		buf, err = ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		p.cache.Add(name, buf)
	}
	return buf, nil
}

// WriteContentType should add the Content-Type header to the response when not set yet.
func (p *Plush2Render) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{p.Options.ContentType}
	}
}
