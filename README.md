# GinTool - Gin Web Framework Config Tool

<img align="right" width="159px" src="https://raw.githubusercontent.com/gin-gonic/logo/master/color.png">

[Gin](https://github.com/gin-gonic/gin) is a web framework written in Go (Golang) with good performance. GinTool target to make Gin config and start much easier.

## Installation

To install GinTool package, you need to install Go and set your Go workspace first.

1. The first need [Go](https://golang.org/) installed (**version 1.11+ is required**), then you can use the below Go command to install GinTool.

```sh
$ go get -u github.com/cytown/gintool
```

2. Import it in your code:

```code
import "github.com/cytown/gintool"
```

3. Test the GinTool

```sh
$ go test
```

4. Start the example

```sh
$ go run example/example.go
```

## Quick start

```go
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/cytown/gintool"
)

func main() {
	ge, err := gintool.NewGin("testdata/gin.conf")
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	ge.Engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"hello": "world"})
	})
	err = ge.Start()
}
```
