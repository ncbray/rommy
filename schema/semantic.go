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

func Resolve(schemas *Schemas) []*rommy.RegionSchema {
	type structWork struct {
		parsed *Struct
		built  *rommy.StructSchema
	}

	type regionWork struct {
		parsed      *Region
		built       *rommy.RegionSchema
		types       map[string]rommy.TypeSchema
		struct_work []structWork
	}

	region_work := []regionWork{}

	// Index
	for _, r := range schemas.Region {
		rr := &rommy.RegionSchema{
			Name: r.Name,
		}

		types := map[string]rommy.TypeSchema{
			"int32":  &rommy.IntegerSchema{},
			"string": &rommy.StringSchema{},
		}

		struct_work := []structWork{}
		for _, s := range r.Struct {
			ss := &rommy.StructSchema{
				Name: s.Name,
			}
			struct_work = append(struct_work, structWork{parsed: s, built: ss})
			types[ss.Name] = ss
			rr.Structs = append(rr.Structs, ss)
		}
		region_work = append(region_work, regionWork{parsed: r, built: rr, types: types, struct_work: struct_work})
	}

	// Resolve types
	for _, rw := range region_work {
		for _, sw := range rw.struct_work {
			s := sw.parsed
			ss := sw.built
			for _, f := range s.Fields {
				ft, ok := getType(rw.types, f.Type)
				if !ok {
					panic(f.Type)
				}
				ss.Fields = append(ss.Fields, &rommy.FieldSchema{
					Name: f.Name,
					Type: ft,
				})
			}
		}
	}

	// Finalize.
	region_list := make([]*rommy.RegionSchema, len(region_work))
	for i, rw := range region_work {
		rw.built.Init()
		region_list[i] = rw.built
	}

	return region_list
}
