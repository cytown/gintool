// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
// The original package is from "github.com/FlowerWrong/plushgin" please respect him

package plushgin

import (
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/plush"
)

// NewContext create a plush.Context
func NewContext(p *Plush2Render, c gin.H) plush.Context {
	pc := *plush.NewContextWith(c)
	for fn, f := range p.helpers {
		pc.Set(fn, f)
	}
	return pc
}
