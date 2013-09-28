// Copyright (c) 2013 Guillaume Delugré.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package selfml

import (
	"errors"
	"io/ioutil"
	"reflect"
)

// Parses a self-ml string and fills the output structure.
func LoadString(data string, out interface{}) (err error) {
	p := selfParser{input: data, r: '\n'}
	rootNode := selfNode{root: true}
	if rootNode.values, err = p.parseNodeBody(true); err != nil {
		return
	}

	println(rootNode.String(0))

	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr && v.Elem().Kind() != reflect.Struct {
		return errors.New("Expected pointer to a struct")
	}

	return rootNode.packToStructByFieldName(v.Elem())
}

// Parses a self-ml file on disk and fills the output structure.
func Load(path string, out interface{}) (err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(path); err != nil {
		return
	}

	return LoadString(string(bytes), out)
}
