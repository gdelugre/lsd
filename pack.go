// Copyright (c) 2013 Guillaume Delugré.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package selfml

import (
	"errors"
	"reflect"
	"strconv"
	"unicode"
)

func publicName(fieldName string) string {
	if len(fieldName) == 0 {
		return fieldName
	} else {
		return string(unicode.ToUpper(rune(fieldName[0]))) + fieldName[1:]
	}
}

func encodeScalarField(name string, kind reflect.Kind, repr string) (interface{}, error) {
	var item interface{}

	switch kind {
	case reflect.String:
		item = repr
	case reflect.Bool:
		if b, err := strconv.ParseBool(repr); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = b
		}
	case reflect.Int:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int(i)
		}
	case reflect.Int8:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int8(i)
		}
	case reflect.Int16:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int16(i)
		}
	case reflect.Int32:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int32(i)
		}
	case reflect.Int64:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int64(i)
		}
	case reflect.Uint:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint(i)
		}
	case reflect.Uint8:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint8(i)
		}
	case reflect.Uint16:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint16(i)
		}
	case reflect.Uint32:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint32(i)
		}
	case reflect.Uint64:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint64(i)
		}
	case reflect.Float32:
		if f, err := strconv.ParseFloat(repr, 32); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = float32(f)
		}
	case reflect.Float64:
		if f, err := strconv.ParseFloat(repr, 64); err != nil {
			return nil, errors.New("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = f
		}
	}

	return item, nil
}

func (node *selfNode) packToStructByFieldName(st reflect.Value) error {

	for _, n := range node.values {
		if _, ok := n.(*selfNode); !ok {
			return errors.New("Field " + node.head.String(0) + " should be only made of lists")
		}
		valueNode := n.(*selfNode)
		fieldName := publicName(valueNode.head.String(0))
		targetField := st.FieldByName(fieldName)
		if !targetField.IsValid() {
			return errors.New("Undefined field " + fieldName)
		}
		fieldKind := targetField.Kind()

		var (
			item interface{}
			err  error
		)

		switch fieldKind {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
			return errors.New("Unsupported field type for field " + fieldName)

		case reflect.Struct:
			if err = valueNode.packToStruct(targetField); err != nil {
				return err
			}

		default:
			if len(valueNode.values) != 1 {
				return errors.New("Bad number of values for field " + fieldName)
			}
			if _, ok := valueNode.values[0].(selfString); !ok {
				return errors.New("Expected a string element for field " + fieldName)
			}
			strValue := valueNode.values[0].(selfString).String(0)

			item, err = encodeScalarField(fieldName, fieldKind, strValue)
			if err != nil {
				return err
			}

			targetField.Set(reflect.ValueOf(item))
		}

		println(fieldName)
	}
	return nil
}

func (node *selfNode) packToStructByFieldOrder(st reflect.Value) error {
	//for i,n := range node.values {
	//	st.Field(i)
	//}
	return nil
}

func (node *selfNode) packToStruct(st reflect.Value) error {

	for _, n := range node.values {
		switch n.(type) {
		case selfString:
			return node.packToStructByFieldOrder(st)
		}
	}
	return node.packToStructByFieldName(st)
}
