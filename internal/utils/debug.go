/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
)

func Dump(i interface{}) {
	var sb strings.Builder
	sb.WriteString("\n") // Start on a new line for better visibility
	dumpRecursive(i, 0, &sb)
	sb.WriteString("== end ==")
	logger.Trace("%s", sb.String())
}

func dumpRecursive(i interface{}, depth int, sb *strings.Builder) {
	if depth > 5 {
		sb.WriteString(fmt.Sprintf("%s[Max Depth Reached]\n", strings.Repeat("  ", depth)))
		return
	}

	v := reflect.ValueOf(i)
	indent := strings.Repeat("  ", depth)

	// Handle pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			sb.WriteString(fmt.Sprintf("%s(nil pointer)\n", indent))
			return
		}
		v = v.Elem()
	}

	// Safety check for invalid values after dereferencing
	if !v.IsValid() {
		sb.WriteString(fmt.Sprintf("%s(invalid)\n", indent))
		return
	}

	t := v.Type()

	// Special Case: Handle Stringers (like time.Time) to avoid unexported field panics
	if v.Kind() == reflect.Struct {
		if stringer, ok := v.Interface().(fmt.Stringer); ok {
			sb.WriteString(fmt.Sprintf("%s%v\n", indent, stringer.String()))
			return
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		sb.WriteString(fmt.Sprintf("%s=== Struct: %s ===\n", indent, t.Name()))
		for i := 0; i < v.NumField(); i++ {
			fName := t.Field(i).Name
			fVal := v.Field(i)

			// Skip unexported fields (e.g., lowercase internal fields)
			if !t.Field(i).IsExported() {
				continue
			}

			if isComplex(fVal) {
				sb.WriteString(fmt.Sprintf("%s%s:\n", indent, fName))
				dumpRecursive(fVal.Interface(), depth+1, sb)
			} else {
				sb.WriteString(fmt.Sprintf("%s%s: %v\n", indent, fName, formatSimpleValue(fVal)))
			}
		}

	case reflect.Map:
		sb.WriteString(fmt.Sprintf("%s=== Map: %s (len %d) ===\n", indent, t.String(), v.Len()))
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			val := iter.Value()

			if isComplex(val) {
				sb.WriteString(fmt.Sprintf("%s[%v]:\n", indent, key))
				dumpRecursive(val.Interface(), depth+1, sb)
			} else {
				sb.WriteString(fmt.Sprintf("%s[%v]: %v\n", indent, key, formatSimpleValue(val)))
			}
		}

	case reflect.Slice, reflect.Array:
		sb.WriteString(fmt.Sprintf("%s=== Slice: %s (len %d) ===\n", indent, t.String(), v.Len()))
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if isComplex(item) {
				sb.WriteString(fmt.Sprintf("%s[%d]:\n", indent, i))
				dumpRecursive(item.Interface(), depth+1, sb)
			} else {
				sb.WriteString(fmt.Sprintf("%s[%d]: %v\n", indent, i, formatSimpleValue(item)))
			}
		}

	default:
		sb.WriteString(fmt.Sprintf("%sValue: %v\n", indent, v.Interface()))
	}
}

func isComplex(v reflect.Value) bool {
	kind := v.Kind()
	if kind == reflect.Ptr && !v.IsNil() {
		kind = v.Elem().Kind()
	}
	// We consider it complex if it's a struct (that isn't a stringer), map, or slice
	if kind == reflect.Struct {
		// If it's a time.Time, it's not "complex" for our purposes because we use .String()
		_, isStringer := v.Interface().(fmt.Stringer)
		return !isStringer
	}
	return kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array
}

func formatSimpleValue(v reflect.Value) string {
	if !v.IsValid() {
		return "invalid"
	}
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		return "nil"
	}
	if !v.CanInterface() {
		return "(unexported)"
	}
	return fmt.Sprintf("%v", v.Interface())
}
