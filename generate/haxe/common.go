package haxe

import (
	"github.com/ncbray/compilerutil/names"
	"github.com/ncbray/rommy/runtime"
)

func structName(s *runtime.StructSchema) string {
	return s.Name
}

func fieldName(f *runtime.FieldSchema) string {
	return names.JoinCamelCase(names.SplitSnakeCase(f.Name), false)
}

func regionName(r *runtime.RegionSchema) string {
	return r.Name + "Region"
}

func poolField(r *runtime.RegionSchema, s *runtime.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Pool"), false)
}

func haxeTypeRef(t runtime.TypeSchema) string {
	switch t := t.(type) {
	case *runtime.IntegerSchema:
		// TODO large integer support
		if t.Bits > 32 {
			panic(t)
		}
		if t.Unsigned {
			return "UInt"
		} else {
			return "Int"
		}
	case *runtime.StringSchema:
		return "String"
	case *runtime.StructSchema:
		return structName(t)
	case *runtime.ListSchema:
		return "Array<" + haxeTypeRef(t.Element) + ">"
	default:
		panic(t)
	}
}
