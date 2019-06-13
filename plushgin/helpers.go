// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
// The original package is from "github.com/FlowerWrong/plushgin" please respect him

package plushgin

func (p *Plush2Render) partial(n string) (string, error) {
	buf, err := p.getCache("_" + n)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (p *Plush2Render) initDefaultHelpers() {
	p.AddHelper("partialFeeder", p.partial)
}

func (p *Plush2Render) AddHelper(fn string, f interface{}) {
	if p.helpers == nil {
		p.helpers = make(map[string]interface{})
	}
	p.helpers[fn] = f
}
