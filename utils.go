package gintool

import "fmt"

func extract(out interface{}, names ...string) (interface{}, error) {
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