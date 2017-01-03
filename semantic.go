package rommy

import (
	"fmt"
	"reflect"
	"strconv"
)

func resolveType(region Region, node Expr, expected TypeSchema, status *Status) (TypeSchema, bool) {
	var loc Location
	var actual TypeSchema
	var ok bool

	switch node := node.(type) {
	case *String:
		loc = node.Raw.Loc
		actual = &StringSchema{}
	case *Integer:
		loc = node.Raw.Loc
		actual = &IntegerSchema{}
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

func reflectionType(t TypeSchema) reflect.Type {
	switch t := t.(type) {
	case *StructSchema:
		return reflect.TypeOf(t.GoType)
	case *ListSchema:
		return reflect.SliceOf(reflectionType(t.Element))
	default:
		panic(t)
	}
}

func handleData(region Region, node Expr, expected TypeSchema, status *Status) (reflect.Value, bool) {
	actual, ok := resolveType(region, node, expected, status)
	if !ok {
		return badValue, false
	}

	switch node := node.(type) {
	case *String:
		_, ok := actual.(*StringSchema)
		if !ok {
			status.Error(node.Raw.Loc, fmt.Sprintf("attempted to instantiate type %s as a string", actual.CanonicalName()))
			return badValue, false
		}
		return reflect.ValueOf(node.Value), true
	case *Integer:
		_, ok := actual.(*IntegerSchema)
		if !ok {
			status.Error(node.Raw.Loc, fmt.Sprintf("attempted to instantiate type %s as an int", actual.CanonicalName()))
			return badValue, false
		}
		// TODO signedness and bits.
		value, err := strconv.ParseInt(node.Raw.Text, 0, 32)
		if err != nil {
			nerr, ok := err.(*strconv.NumError)
			if !ok {
				panic(nerr)
			}
			if nerr.Err != strconv.ErrRange {
				panic(err)
			}
			status.Error(node.Raw.Loc, fmt.Sprintf("%s out of range for an int32", node.Raw.Text))
			return badValue, false
		}
		return reflect.ValueOf(int32(value)), true
	case *Struct:
		t, ok := actual.(*StructSchema)
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
		t, ok := expected.(*ListSchema)
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

type Region interface {
	Schema() *RegionSchema
	Allocate(name string) interface{}
}

func HandleData(region Region, node Expr, expected TypeSchema, status *Status) (interface{}, bool) {
	rv, ok := handleData(region, node, expected, status)
	if ok {
		return rv.Interface(), true
	} else {
		return nil, false
	}
}

// This function can be used if there is only one data file to parse.
func ParseFile(file string, data []byte, region Region) (interface{}, bool) {
	sources := CreateSourceSet()
	status := &Status{Sources: sources}
	info := sources.Add(file, data)
	e := ParseData(info, data, status)
	if status.ShouldStop() {
		return nil, false
	}
	return HandleData(region, e, nil, status)
}
