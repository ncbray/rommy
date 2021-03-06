// Package human handles human-readable text data files.
package human

import (
	"fmt"
	"github.com/ncbray/rommy/parser"
	"github.com/ncbray/rommy/runtime"
	"reflect"
	"strconv"
)

func resolveType(region runtime.Region, node Expr, expected runtime.TypeSchema, status *parser.Status) (runtime.TypeSchema, bool) {
	var loc parser.Location
	var actual runtime.TypeSchema
	var ok bool

	switch node := node.(type) {
	case *String:
		loc = node.Raw.Loc
		actual = &runtime.StringSchema{}
	case *Boolean:
		loc = node.Loc
		actual = &runtime.BooleanSchema{}
	case *Integer:
		var bits uint8 = 64
		unsigned := false
		e, ok := expected.(*runtime.IntegerSchema)
		if ok {
			bits = e.Bits
			unsigned = e.Unsigned
		}
		loc = node.Raw.Loc
		actual = &runtime.IntegerSchema{Bits: bits, Unsigned: unsigned}
	case *Struct:
		if node.Type != nil {
			type_name := node.Type.Raw
			loc = type_name.Loc
			rs := region.Schema()
			actual, ok = rs.StructLUT[type_name.Text]
			if !ok {
				status.Error(type_name.Loc, fmt.Sprintf("cannot resolve type %#v", type_name.Text))
				return nil, false
			}
		} else {
			loc = node.Loc
		}
	case *List:
		loc = node.Loc
	default:
		panic(node)
	}

	if actual != nil {
		if expected != nil {
			if !expected.CanHold(actual) {
				status.Error(loc, fmt.Sprintf("expected type %s, but got type %s", expected.CanonicalName(), actual.CanonicalName()))
				return nil, false
			}
		}
	} else {
		if expected != nil {
			actual = expected
		} else {
			status.Error(loc, "cannot determine type")
			return nil, false
		}
	}

	return actual, true
}

var badValue = reflect.ValueOf(nil)

func reflectionType(t runtime.TypeSchema) reflect.Type {
	switch t := t.(type) {
	case *runtime.StructSchema:
		return reflect.TypeOf(t.GoType)
	case *runtime.ListSchema:
		return reflect.SliceOf(reflectionType(t.Element))
	case *runtime.StringSchema:
		return reflect.TypeOf("")
	case *runtime.BooleanSchema:
		return reflect.TypeOf(false)
	default:
		panic(t)
	}
}

func handleIntParseError(value parser.SourceString, err error, t *runtime.IntegerSchema, status *parser.Status) {
	nerr, ok := err.(*strconv.NumError)
	if !ok {
		panic(nerr)
	}
	if nerr.Err != strconv.ErrRange {
		panic(err)
	}
	status.Error(value.Loc, fmt.Sprintf("%s out of range for an %s", value.Text, t.CanonicalName()))
}

func handleData(region runtime.Region, node Expr, expected runtime.TypeSchema, status *parser.Status) (reflect.Value, bool) {
	actual, ok := resolveType(region, node, expected, status)
	if !ok {
		return badValue, false
	}

	switch node := node.(type) {
	case *String:
		_, ok := actual.(*runtime.StringSchema)
		if !ok {
			status.Error(node.Raw.Loc, fmt.Sprintf("attempted to instantiate type %s as a string", actual.CanonicalName()))
			return badValue, false
		}
		return reflect.ValueOf(node.Value), true
	case *Boolean:
		_, ok := actual.(*runtime.BooleanSchema)
		if !ok {
			status.Error(node.Loc, fmt.Sprintf("attempted to instantiate type %s as a bool", actual.CanonicalName()))
			return badValue, false
		}
		return reflect.ValueOf(node.Value), true
	case *Integer:
		t, ok := actual.(*runtime.IntegerSchema)
		if !ok {
			status.Error(node.Raw.Loc, fmt.Sprintf("attempted to instantiate type %s as an int", actual.CanonicalName()))
			return badValue, false
		}
		if t.Unsigned {
			value, err := strconv.ParseUint(node.Raw.Text, 0, int(t.Bits))
			if err != nil {
				handleIntParseError(node.Raw, err, t, status)
				return badValue, false
			}
			switch t.Bits {
			case 8:
				return reflect.ValueOf(uint8(value)), true
			case 16:
				return reflect.ValueOf(uint16(value)), true
			case 32:
				return reflect.ValueOf(uint32(value)), true
			case 64:
				return reflect.ValueOf(uint64(value)), true
			default:
				panic(t.CanonicalName())
			}
		} else {
			value, err := strconv.ParseInt(node.Raw.Text, 0, int(t.Bits))
			if err != nil {
				handleIntParseError(node.Raw, err, t, status)
				return badValue, false
			}
			switch t.Bits {
			case 8:
				return reflect.ValueOf(int8(value)), true
			case 16:
				return reflect.ValueOf(int16(value)), true
			case 32:
				return reflect.ValueOf(int32(value)), true
			case 64:
				return reflect.ValueOf(int64(value)), true
			default:
				panic(t.CanonicalName())
			}
		}
	case *Struct:
		t, ok := actual.(*runtime.StructSchema)
		if !ok {
			status.Error(node.Loc, fmt.Sprintf("attempted to instantiate type %s as a struct", actual.CanonicalName()))
			return badValue, false
		}
		inst := region.Allocate(t.Name)
		if inst == nil {
			panic(inst)
		}
		rv := reflect.ValueOf(inst)

		all_ok := true
		defined := make([]bool, len(t.Fields))
		for _, arg := range node.Args {
			f, ok := t.FieldLUT[arg.Name.Text]
			if ok {
				if defined[f.ID] {
					status.Error(arg.Name.Loc, fmt.Sprintf("attempted to re-define %#v", arg.Name.Text))
				} else {
					defined[f.ID] = true
				}
				fv, ok := handleData(region, arg.Value, f.Type, status)
				if ok {
					rf := rv.Elem().FieldByName(f.GoName())
					rf.Set(fv)
				} else {
					all_ok = false
				}
			} else {
				status.Error(arg.Name.Loc, fmt.Sprintf("type %s does not have field %#v", t.CanonicalName(), arg.Name.Text))
				all_ok = false
			}
		}

		if all_ok {
			return rv, true
		} else {
			return badValue, false
		}
	case *List:
		t, ok := expected.(*runtime.ListSchema)
		if !ok {
			status.Error(node.Loc, fmt.Sprintf("attempted to instantiate type %s as a list", expected.CanonicalName()))
			return badValue, false
		}
		rt := reflectionType(t)
		rv := reflect.MakeSlice(rt, len(node.Args), len(node.Args))
		all_ok := true
		for i, arg := range node.Args {
			fv, ok := handleData(region, arg, t.Element, status)
			if ok {
				rf := rv.Index(i)
				rf.Set(fv)
			} else {
				all_ok = false
			}
		}
		if all_ok {
			return rv, true
		} else {
			return badValue, false
		}
	default:
		panic(node)
	}
}

// Convert an AST to a runtime structure.
func DataToStruct(region runtime.Region, node Expr, expected runtime.TypeSchema, status *parser.Status) (runtime.Struct, bool) {
	rv, ok := handleData(region, node, expected, status)
	if ok {
		general := rv.Interface()
		specific, ok := general.(runtime.Struct)
		if !ok {
			panic(general)
		}
		return specific, true
	} else {
		return nil, false
	}
}

// Simple interface for parsing a single data file.
func ParseFile(file string, data []byte, region runtime.Region) (runtime.Struct, bool) {
	sources := parser.CreateSourceSet()
	status := &parser.Status{Sources: sources}
	info := sources.Add(file, data)
	e := ParseData(info, data, status)
	if status.ShouldStop() {
		return nil, false
	}
	return DataToStruct(region, e, nil, status)
}
