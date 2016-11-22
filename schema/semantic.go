package schema

import (
	"github.com/ncbray/rommy"
)

func getType(types map[string]rommy.TypeSchema, name string) (rommy.TypeSchema, bool) {
	if len(name) >= 2 && name[0:2] == "[]" {
		t, ok := getType(types, name[2:])
		if ok {
			return t.List(), true
		} else {
			return nil, false
		}
	}
	t, ok := types[name]
	return t, ok
}

func Resolve(schemas *Schemas) []*rommy.StructSchema {
	types := map[string]rommy.TypeSchema{
		"int32":  &rommy.IntegerSchema{},
		"string": &rommy.StringSchema{},
	}

	struct_list := []*rommy.StructSchema{}

	// Index
	for _, s := range schemas.Struct {
		ss := &rommy.StructSchema{
			Name: s.Name,
		}
		struct_list = append(struct_list, ss)
		types[ss.Name] = ss
	}

	// Resolve types
	for i, s := range schemas.Struct {
		ss := struct_list[i]
		for _, f := range s.Fields {
			ft, ok := getType(types, f.Type)
			if !ok {
				panic(f.Type)
			}
			ss.Fields = append(ss.Fields, &rommy.FieldSchema{
				Name: f.Name,
				Type: ft,
			})
		}
	}

	// Finalize.
	for _, ss := range struct_list {
		ss.Init()
	}

	return struct_list
}
