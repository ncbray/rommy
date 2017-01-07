package golang

import (
	"github.com/ncbray/compilerutil/names"
	"github.com/ncbray/rommy/runtime"
)

func goTypeRef(t runtime.TypeSchema) string {
	switch t := t.(type) {
	case *runtime.IntegerSchema:
		// HACK canonical name matches Go type.
		return t.CanonicalName()
	case *runtime.StringSchema:
		return "string"
	case *runtime.BooleanSchema:
		return "bool"
	case *runtime.StructSchema:
		return "*" + t.Name
	case *runtime.ListSchema:
		return "[]" + goTypeRef(t.Element)
	default:
		panic(t)
	}
}

func fieldName(f *runtime.FieldSchema) string {
	return names.JoinCamelCase(names.SplitSnakeCase(f.Name), true)
}

func structSchemaName(s *runtime.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Schema"), false)
}

func regionStructName(r *runtime.RegionSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(r.Name+"Region"), true)
}

func regionSchemaName(r *runtime.RegionSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(r.Name+"RegionSchema"), false)
}

func poolField(r *runtime.RegionSchema, s *runtime.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Pool"), true)
}

func regionClonerName(r *runtime.RegionSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(r.Name+"Cloner"), true)
}

func mapingField(r *runtime.RegionSchema, s *runtime.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Map"), false)
}
