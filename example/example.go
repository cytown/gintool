// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/v2pro/plz/gls"

	"github.com/cytown/gintool"
)

func main() {
	ge, err := gintool.NewGin("testdata/gin.conf")
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	c := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		//log.Printf("captured %v, stopping profiler and exiting..", sig)
		fmt.Println("shutdown return", ge.ShutDown())
		done <- true
	}()
	log.Println("ginengine: ", ge)
	ge.Engine.Use(func(c *gin.Context) {
		defer func() {
			log.Println("test middle ware end", gls.GoID())
		}()
		//log.Println("test middle ware start")
		gintool.SessionSet("test", "great")
		log.Println("test middle ware start2", gls.GoID(), gintool.SessionGet("test"))
		c.Next()
	})
	ge.Engine.GET("/", func(c *gin.Context) {
		log.Println("session key test", gls.GoID(), gintool.SessionGet("test"))
		panic("test")
		c.JSON(200, []string{"test", gintool.SessionGet("test").(string)})
	})
	err = ge.Start()
	if err != nil && err != http.ErrServerClosed {
		fmt.Println("error2: ", err)
	} else {
		<-done
	}
}
