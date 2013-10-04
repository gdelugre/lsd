// Copyright (c) 2013 Guillaume Delugré.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package lsd

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Error type that can be triggered while packing values.
type packError struct {
	message    string
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

// Checks whether a kind can be packed as a single scalar value.
func isScalarKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128,
		reflect.Struct, reflect.Slice, reflect.Array:
		return false

	default:
		return true
	}
}

// Checks whether a kind represents a compound type.
func isCompoundKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:
		return true
	default:
		return false
	}
}

// Allowed node heads to be used as bullet points.
func isBulletPoint(str string) bool {
	r, _ := utf8.DecodeRuneInString(str)
	return r == '-' || r == '*' || r == '•' || r == '◦' || r == '‣' || r == '⁃'
}

// Extended version of strconv.ParseInt.
// Also accepts binary forms with "0b" prefix.
func parseIntEx(s string, bitSize int) (int64, error) {
	if s[0:2] == "0b" {
		return strconv.ParseInt(s[2:], 2, bitSize)
	} else {
		return strconv.ParseInt(s, 0, bitSize)
	}
}

// Extended version of strconv.ParseUint.
// Also accepts binary forms with "0b" prefix.
func parseUintEx(s string, bitSize int) (uint64, error) {
	if s[0:2] == "0b" {
		return strconv.ParseUint(s[2:], 2, bitSize)
	} else {
		return strconv.ParseUint(s, 0, bitSize)
	}
}

// Extended version of strconv.ParseBool.
// Also accepts variations of "Yes" and "No" strings.
func parseBoolEx(repr string) (value bool, err error) {
	if value, err = strconv.ParseBool(repr); err != nil {
		switch repr {
		case "y", "yes", "YES", "Yes":
			return true, nil
		case "n", "no", "NO", "No":
			return false, nil
		}
	}

	return
}

// Converts a string to its native non-compound Go type.
func (str selfString) encodeScalarField(kind reflect.Kind) (interface{}, error) {
	var item interface{}

	repr := str.String()
	switch kind {
	case reflect.String:
		item = repr
	case reflect.Bool:
		if b, err := parseBoolEx(repr); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = b
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var bitSize int
		switch kind {
		case reflect.Int:
			bitSize = 0
		case reflect.Int8:
			bitSize = 8
		case reflect.Int16:
			bitSize = 16
		case reflect.Int32:
			bitSize = 32
		case reflect.Int64:
			bitSize = 64
		}
		if i, err := parseIntEx(repr, bitSize); err != nil {
			return nil, str.newPackError("cannot convert value `" + repr + "` to type " + kind.String())
		} else {
			switch kind {
			case reflect.Int:
				item = int(i)
			case reflect.Int8:
				item = int8(i)
			case reflect.Int16:
				item = int16(i)
			case reflect.Int32:
				item = int32(i)
			case reflect.Int64:
				item = int64(i)
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var bitSize int
		switch kind {
		case reflect.Uint:
			bitSize = 0
		case reflect.Uint8:
			bitSize = 8
		case reflect.Uint16:
			bitSize = 16
		case reflect.Uint32:
			bitSize = 32
		case reflect.Uint64:
			bitSize = 64
		}
		if u, err := parseUintEx(repr, bitSize); err != nil {
			return nil, str.newPackError("cannot convert value `" + repr + "` to type " + kind.String())
		} else {
			switch kind {
			case reflect.Uint:
				item = uint(u)
			case reflect.Uint8:
				item = uint8(u)
			case reflect.Uint16:
				item = uint16(u)
			case reflect.Uint32:
				item = uint32(u)
			case reflect.Uint64:
				item = uint64(u)
			}
		}

	case reflect.Float32:
		if f, err := strconv.ParseFloat(repr, 32); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = float32(f)
		}
	case reflect.Float64:
		if f, err := strconv.ParseFloat(repr, 64); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = f
		}
	}

	return item, nil
}

// Packs a selfNode into a Go structure/map field.
// If fhe field is a scalar type, process it with encodeScalarField.
// If the field is a structure, process it with packToStruct.
func (node selfNode) packIntoField(name string, field reflect.Value) (err error) {

	fieldKind := field.Kind()

	if isScalarKind(fieldKind) {
		if len(node.values) != 1 {
			return node.newPackError("bad number of values for scalar field `" + name + "`")
		}
		if _, ok := node.values[0].(selfString); !ok {
			return node.newPackError("expected a string element for scalar field `" + name + "`")
		}
		strValue := node.values[0].(selfString)
		return strValue.packIntoField(name, field)

	} else if fieldKind == reflect.Struct {
		return node.packToStruct(field)

	} else if fieldKind == reflect.Array {
		return node.packToArray(field)

	} else if fieldKind == reflect.Slice {
		return node.packToSlice(field)

	} else if fieldKind == reflect.Map {
		field.Set(reflect.MakeMap(field.Type())) // Map requires initialization.
		return node.packToMap(field)

	} else {
		return node.newPackError("unsupported field kind " + fieldKind.String())
	}

	return
}

// Packs a selfString into a Go structure/map field.
// The field type must be scalar to hold the value.
func (str selfString) packIntoField(_ string, field reflect.Value) (err error) {

	var value reflect.Value
	if value, err = str.makeValue(field.Type()); err != nil {
		return
	}

	field.Set(value)
	return
}

// Packs a selfString into a new allocated reflect.Value.
// This value can later be set into a field or variable.
func (str selfString) makeValue(t reflect.Type) (value reflect.Value, err error) {

	var item interface{}
	kind := t.Kind()
	value = reflect.Zero(t)

	if isScalarKind(kind) {
		if item, err = str.encodeScalarField(kind); err != nil {
			return
		}
		value = reflect.ValueOf(item)

	} else if isCompoundKind(kind) {
		err = str.newPackError("cannot pack string `" + str.String() + "` into field of compound kind " + kind.String())

	} else {
		err = str.newPackError("unsupported field kind " + kind.String())
	}

	return
}

// Packs a selfNode into a new allocated reflect.Value.
// This value can later be set into a field or variable.
func (node selfNode) makeValue(t reflect.Type) (value reflect.Value, err error) {

	kind := t.Kind()
	value = reflect.Zero(t)

	if isScalarKind(kind) {
		err = node.newPackError("expected a string element for scalar field")

	} else if kind == reflect.Array {
		value = reflect.MakeSlice(t, t.Len(), t.Len())
		err = node.packToArray(value)

	} else if kind == reflect.Slice {
		value = reflect.New(t).Elem()
		err = node.packToSlice(value)

	} else if kind == reflect.Struct {
		value = reflect.New(t).Elem()
		err = node.packToStruct(value)

	} else if kind == reflect.Map {
		value = reflect.MakeMap(t)
		err = node.packToMap(value)

	} else {
		err = node.newPackError("unsupported field kind " + kind.String())
	}

	return
}

// Check the header value for a compound type nested into a slice or array.
func (node *selfNode) checkMetaHeader(elemType reflect.Type) error {

	header := node.head.String()
	kind := elemType.Kind()

	if kind == reflect.Slice || kind == reflect.Array {
		// Packing a slice of slices requires the [] (empty string) header.
		if len(header) != 0 {
			return node.head.newPackError("slice head has value `" + header + "` instead of []")
		}

	} else if kind == reflect.Struct || kind == reflect.Map {
		// Packing a slice of structs or maps. Requires the type name as header or a bullet point.
		if !isBulletPoint(header) && header != elemType.Name() {
			return node.head.newPackError("struct head has value `" + header + "` instead of bullet or `" + elemType.Name() + "`")
		}
	}

	return nil
}

// Packs a selfNode into a Go array.
func (node *selfNode) packToArray(field reflect.Value) (err error) {

	arraySize := field.Type().Len()
	if len(node.values) > arraySize {
		return node.newPackError(fmt.Sprintf("too many values to fit into array of %d elements", arraySize))
	}

	arrayType := field.Type().Elem()
	arrayKind := arrayType.Kind()

	for i, n := range node.values {

		switch arrayKind {
		case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
			if _, ok := n.(*selfNode); !ok {
				return n.newPackError("compound kind `" + arrayKind.String() + "` expected a list of values")
			}

			subNode := n.(*selfNode)
			if err = subNode.checkMetaHeader(arrayType); err != nil {
				return
			}
		}

		if err = n.packIntoField("", field.Index(i)); err != nil {
			return
		}
	}

	return
}

// Packs a selfNode into a Go slice.
func (node *selfNode) packToSlice(field reflect.Value) (err error) {
	sliceType := field.Type().Elem()
	sliceKind := sliceType.Kind()

	var value reflect.Value
	for _, n := range node.values {

		switch sliceKind {
		case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
			if _, ok := n.(*selfNode); !ok {
				return n.newPackError("compound kind `" + sliceKind.String() + "` expected a list of values")
			}

			subNode := n.(*selfNode)
			if err = subNode.checkMetaHeader(sliceType); err != nil {
				return
			}
		}

		if value, err = n.makeValue(sliceType); err != nil {
			return
		}

		field.Set(reflect.Append(field, value))
	}

	return nil
}

// Packs a selfNode into a Go map.
// Values must be nodes as their heads are used as keys into the map.
func (node *selfNode) packToMap(m reflect.Value) (err error) {

	var (
		key   interface{}
		value reflect.Value
	)

	nodeName := node.head.String()
	keyType, elemType := m.Type().Key(), m.Type().Elem()

	for _, n := range node.values {
		if _, ok := n.(*selfNode); !ok {
			return n.newPackError("field `" + nodeName + "` should be only made of lists")
		}
		valueNode := n.(*selfNode)
		nodeHead := valueNode.head
		if key, err = nodeHead.encodeScalarField(keyType.Kind()); err != nil {
			return
		}

		value = reflect.New(elemType).Elem()
		if err = valueNode.packIntoField(nodeHead.String(), value); err != nil {
			return
		}

		m.SetMapIndex(reflect.ValueOf(key), value)
	}
	return
}

// Packs a selfNode into a Go structure.
// For each iterated member in the node, fills the corresponding structure field by name.
func (node *selfNode) packToStructByFieldName(st reflect.Value) (err error) {

	nodeName := node.head.String()
	for _, n := range node.values {
		if _, ok := n.(*selfNode); !ok {
			return n.newPackError("field `" + nodeName + "` should be only made of lists")
		}
		valueNode := n.(*selfNode)
		fieldName := publicName(valueNode.head.String())
		targetField := st.FieldByName(fieldName)
		if !targetField.IsValid() {
			return valueNode.newPackError("undefined field `" + fieldName + "` for node `" + nodeName + "`")
		}

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
		return node.newPackError("too many values to fit into struct " + typeName)
	}

	nodeName := node.head.String()
	if !isBulletPoint(nodeName) && typeName != nodeName {
		return node.newPackError("bad value `" + nodeName + "`, expected `" + typeName + "`")
	}

	for i, n := range node.values {
		targetField := st.Field(i)
		if err = n.packIntoField("", targetField); err != nil {
			return
		}
	}
	return nil
}

// Packs a selfNode into a Go structure.
// If the node only contains subnodes and their heads match field names, consider filling each field by name.
func (node *selfNode) packToStruct(st reflect.Value) error {

	for _, n := range node.values {
		switch n.(type) {
		case selfString:
			return node.packToStructByFieldOrder(st)

		case *selfNode:
			if !st.FieldByName(n.(*selfNode).head.String()).IsValid() {
				return node.packToStructByFieldOrder(st)
			}
		}
	}
	return node.packToStructByFieldName(st)
}
