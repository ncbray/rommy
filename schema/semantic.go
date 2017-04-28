package schema

import (
	"github.com/ncbray/rommy/runtime"
)

func getType(types map[string]runtime.TypeSchema, name string) (runtime.TypeSchema, bool) {
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

func Resolve(schemas *Schemas) []*runtime.RegionSchema {
	type structWork struct {
		parsed *Struct
		built  *runtime.StructSchema
	}

	type regionWork struct {
		parsed      *Region
		built       *runtime.RegionSchema
		types       map[string]runtime.TypeSchema
		struct_work []structWork
	}

	region_work := []regionWork{}

	// Index
	for _, r := range schemas.Region {
		rr := &runtime.RegionSchema{
			Name: r.Name,
		}

		types := map[string]runtime.TypeSchema{
			"string": &runtime.StringSchema{},
			"bool":   &runtime.BooleanSchema{},
		}
		for _, unsigned := range []bool{false, true} {
			for _, bits := range []uint8{8, 16, 32, 64} {
				t := &runtime.IntegerSchema{Bits: bits, Unsigned: unsigned}
				types[t.CanonicalName()] = t
			}
		}

		for _, bits := range []uint8{32, 64} {
			t := &runtime.FloatSchema{Bits: bits}
			types[t.CanonicalName()] = t
		}

		struct_work := []structWork{}
		for _, s := range r.Struct {
			ss := &runtime.StructSchema{
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
					panic("cannot resolve type " + f.Type)
				}
				ss.Fields = append(ss.Fields, &runtime.FieldSchema{
					Name: f.Name,
					Type: ft,
				})
			}
		}
	}

	// Finalize.
	region_list := make([]*runtime.RegionSchema, len(region_work))
	for i, rw := range region_work {
		rw.built.Init()
		region_list[i] = rw.built
	}

	return region_list
}
