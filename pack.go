// Copyright (c) 2013 Guillaume Delugré.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package selfml

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"
)

// Error type that can be triggered while packing values.
type packError struct {
	message string
	lineNumber uint
}

// Generates an error while packing a selfString.
func (v selfString) newPackError(str string) error {
	return &packError{message: str, lineNumber: v.LineNumber()}
}

// Generates an error while packing a selfNode.
func (v selfNode) newPackError(str string) error {
	return &packError{message: str, lineNumber: v.LineNumber()}
}

// Error printing.
func (err packError) Error() (str string) {
	str = fmt.Sprintf("Error while packing structure: %s", err.message)
	if err.lineNumber != 0 {
		str += fmt.Sprintf(" (line %d)", err.lineNumber)
	}
	return
}

// Gets the line number where a node was defined.
func (node selfNode) LineNumber() uint {
	return node.lineNumber
}

// Gets the line number where a string value was defined.
func (str selfString) LineNumber() uint {
	return str.lineNumber
}

// Capitalize the name of a structure field.
func publicName(fieldName string) string {
	if len(fieldName) == 0 {
		return fieldName
	} else {
		return string(unicode.ToUpper(rune(fieldName[0]))) + fieldName[1:]
	}
}

// Converts a string to its native non-compound Go type.
func (str selfString) encodeScalarField(name string, kind reflect.Kind) (interface{}, error) {
	var item interface{}

	repr := str.String()
	switch kind {
	case reflect.String:
		item = repr
	case reflect.Bool:
		if b, err := strconv.ParseBool(repr); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = b
		}
	case reflect.Int:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int(i)
		}
	case reflect.Int8:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int8(i)
		}
	case reflect.Int16:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int16(i)
		}
	case reflect.Int32:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int32(i)
		}
	case reflect.Int64:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = int64(i)
		}
	case reflect.Uint:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint(i)
		}
	case reflect.Uint8:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint8(i)
		}
	case reflect.Uint16:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint16(i)
		}
	case reflect.Uint32:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint32(i)
		}
	case reflect.Uint64:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = uint64(i)
		}
	case reflect.Float32:
		if f, err := strconv.ParseFloat(repr, 32); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = float32(f)
		}
	case reflect.Float64:
		if f, err := strconv.ParseFloat(repr, 64); err != nil {
			return nil, str.newPackError("Cannot convert field " + name + " to type " + kind.String())
		} else {
			item = f
		}
	}

	return item, nil
}

func (node selfNode) packIntoField(name string, field reflect.Value) (err error) {
	var item interface{}

	fieldKind := field.Kind()

	switch fieldKind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
		return node.newPackError("Unsupported field kind: " + fieldKind.String())

	case reflect.Struct:
		if err = node.packToStruct(field); err != nil {
			return
		}

	case reflect.Slice:
		sliceType := field.Type().Elem()
		println(sliceType.String())
		return node.newPackError("fuckyou slice")

	default:
		if len(node.values) != 1 {
			return node.newPackError("Bad number of values for scalar field " + name)
		}
		if _, ok := node.values[0].(selfString); !ok {
			return node.newPackError("Expected a string element for scalar field " + name)
		}
		strValue := node.values[0].(selfString)

		if item, err = strValue.encodeScalarField(name, fieldKind); err != nil {
			return
		}

		field.Set(reflect.ValueOf(item))
	}

	return
}

func (str selfString) packIntoField(name string, field reflect.Value) (err error) {
	var item interface{}
	fieldKind := field.Kind()

	switch fieldKind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
		return str.newPackError("Unsupported field kind: " + fieldKind.String())

	case reflect.Struct, reflect.Slice, reflect.Array:
		return str.newPackError("Cannot pack string `" + str.String() + "` into field of compound kind " + fieldKind.String())

	default:
		if item, err = str.encodeScalarField(name, fieldKind); err != nil {
			return
		}
		field.Set(reflect.ValueOf(item))
	}
	return
}

// Packs a selfNode into a Go structure.
// For each iterated member in the node, fills the corresponding structure field by name.
func (node *selfNode) packToStructByFieldName(st reflect.Value) (err error) {

	for _, n := range node.values {
		nodeName := node.head.String()
		if _, ok := n.(*selfNode); !ok {
			return node.newPackError("Field `" + nodeName + "` should be only made of lists")
		}
		valueNode := n.(*selfNode)
		fieldName := publicName(valueNode.head.String())
		targetField := st.FieldByName(fieldName)
		if !targetField.IsValid() {
			return valueNode.newPackError("Undefined field `" + fieldName + "` for node `" + nodeName + "`")
		}

		println(fieldName)
		if err = valueNode.packIntoField(fieldName, targetField); err != nil {
			return
		}
	}
	return nil
}

// Packs a selfNode into a Go structure.
// For each iterated member in the node, fills the corresponding structure field by order.
// Node head must match with structure type name (empty string for anonymous struct).
func (node *selfNode) packToStructByFieldOrder(st reflect.Value) (err error) {

	typeName := st.Type().Name()
	if st.NumField() < len(node.values) {
		return node.newPackError("Too many values to fit into struct " + typeName)
	}

	nodeName := node.head.String()
	if typeName != nodeName {
		return node.newPackError("Bad value `" + nodeName + "`, expected `" + typeName + "`")
	}

	for i,n := range node.values {
		targetField := st.Field(i)
		if err = n.packIntoField("", targetField); err != nil {
			return
		}
	}
	return nil
}

// Packs a selfNode into a Go structure.
// If the node only contains subnodes, consider filling each field by name.
func (node *selfNode) packToStruct(st reflect.Value) error {

	for _, n := range node.values {
		switch n.(type) {
		case selfString:
			return node.packToStructByFieldOrder(st)
		}
	}
	return node.packToStructByFieldName(st)
}
