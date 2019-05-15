// Copyright 2019 Cytown.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gintool

import (
	"fmt"
	"os"
)

func extract(out interface{}, names ...string) (interface{}, error) {
	if out == nil {
		return nil, fmt.Errorf("can't extract nil")
	}
	var tmp, ok = out.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("%v is not a map", out)
	}
	for i, name := range names {
		//fmt.Println("parse", name, "in", tmp)
		t1, ok := tmp[name]
		if !ok {
			return nil, fmt.Errorf("not found [%v] in map", name)
		}
		if i == len(names)-1 {
			return t1, nil
		}
		tmp, ok = t1.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("[%v] is not a map", name)
		}
	}
	return nil, nil
}

func errorName(code int) string {
	return fmt.Sprintf("__E_%d", code)
}

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
