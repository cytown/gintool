// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
// The original package is from "github.com/FlowerWrong/plushgin" please respect him

package plushgin

import (
	lru "github.com/hashicorp/golang-lru"
)

// Some package-internal structure helping to implement render cache
type templateCache struct {
	cache *lru.Cache
}

func newTemplateCache(max int) *templateCache {
	arcCache, _ := lru.New(max)

	return &templateCache{
		cache: arcCache,
	}
}

func (c *templateCache) Get(templateName string) []byte {
	rendered, alreadyInCache := c.cache.Get(templateName)

	if alreadyInCache {
		return rendered.([]byte)
	}
	return nil
}

func (c *templateCache) Add(templateName string, content []byte) {
	c.cache.Add(templateName, content)
}
