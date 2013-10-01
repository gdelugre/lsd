// Copyright (c) 2013 Guillaume Delugré.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package selfml

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
	case reflect.Array, reflect.Slice, reflect.Struct:
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
	case reflect.Int:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = int(i)
		}
	case reflect.Int8:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = int8(i)
		}
	case reflect.Int16:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = int16(i)
		}
	case reflect.Int32:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = int32(i)
		}
	case reflect.Int64:
		if i, err := strconv.Atoi(repr); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = int64(i)
		}
	case reflect.Uint:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = uint(i)
		}
	case reflect.Uint8:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = uint8(i)
		}
	case reflect.Uint16:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = uint16(i)
		}
	case reflect.Uint32:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = uint32(i)
		}
	case reflect.Uint64:
		if i, err := strconv.ParseUint(repr, 10, 0); err != nil {
			return nil, str.newPackError("cannot convert value `" + str.String() + "` to type " + kind.String())
		} else {
			item = uint64(i)
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

// Packs a selfNode into a Go slice.
func (node selfNode) packToSlice(field reflect.Value) (err error) {
	sliceType := field.Type().Elem()
	sliceKind := sliceType.Kind()

	var value reflect.Value
	for _, n := range node.values {

		if sliceKind == reflect.Slice {
			// Packing a slice of slices requires the [] (empty string) header.
			if _, ok := n.(*selfNode); !ok {
				return n.newPackError("slice type expected a list of values")
			}
			subNode := n.(*selfNode)
			if len(subNode.head.String()) != 0 {
				return subNode.head.newPackError("slice head has value `" + subNode.head.String() + "` instead of []")
			}

		} else if sliceKind == reflect.Struct || sliceKind == reflect.Map {
			// Packing a slice of structs or maps. Requires the type name as header or a bullet point.
			if _, ok := n.(*selfNode); !ok {
				return n.newPackError("struct/map type expected a list of values")
			}
			subNode := n.(*selfNode)
			subHead := subNode.head.String()
			if !isBulletPoint(subHead) && subHead != sliceType.Name() {
				return subNode.head.newPackError("struct head has value `" + subHead + "` instead of bullet or `" + sliceType.Name() + "`")
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
func (node selfNode) packToMap(m reflect.Value) (err error) {

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
