package schema

/* Generated with rommygen, do not edit by hand. */

import (
	"github.com/ncbray/rommy/runtime"
)

type Field struct {
	PoolIndex int
	Name      string
	Type      string
}

func (s *Field) Schema() *runtime.StructSchema {
	return fieldSchema
}

var fieldSchema = &runtime.StructSchema{Name: "Field", GoType: (*Field)(nil)}

type Struct struct {
	PoolIndex int
	Name      string
	Fields    []*Field
}

func (s *Struct) Schema() *runtime.StructSchema {
	return structSchema
}

var structSchema = &runtime.StructSchema{Name: "Struct", GoType: (*Struct)(nil)}

type Region struct {
	PoolIndex int
	Name      string
	Struct    []*Struct
}

func (s *Region) Schema() *runtime.StructSchema {
	return regionSchema
}

var regionSchema = &runtime.StructSchema{Name: "Region", GoType: (*Region)(nil)}

type Schemas struct {
	PoolIndex int
	Region    []*Region
}

func (s *Schemas) Schema() *runtime.StructSchema {
	return schemasSchema
}

var schemasSchema = &runtime.StructSchema{Name: "Schemas", GoType: (*Schemas)(nil)}

type TypeDeclRegion struct {
	FieldPool   []*Field
	StructPool  []*Struct
	RegionPool  []*Region
	SchemasPool []*Schemas
}

func (r *TypeDeclRegion) Schema() *runtime.RegionSchema {
	return typeDeclRegionSchema
}

func (r *TypeDeclRegion) AllocateField() *Field {
	o := &Field{}
	o.PoolIndex = len(r.FieldPool)
	r.FieldPool = append(r.FieldPool, o)
	return o
}

func (r *TypeDeclRegion) AllocateStruct() *Struct {
	o := &Struct{}
	o.PoolIndex = len(r.StructPool)
	r.StructPool = append(r.StructPool, o)
	return o
}

func (r *TypeDeclRegion) AllocateRegion() *Region {
	o := &Region{}
	o.PoolIndex = len(r.RegionPool)
	r.RegionPool = append(r.RegionPool, o)
	return o
}

func (r *TypeDeclRegion) AllocateSchemas() *Schemas {
	o := &Schemas{}
	o.PoolIndex = len(r.SchemasPool)
	r.SchemasPool = append(r.SchemasPool, o)
	return o
}

func (r *TypeDeclRegion) Allocate(name string) interface{} {
	switch name {
	case "Field":
		return r.AllocateField()
	case "Struct":
		return r.AllocateStruct()
	case "Region":
		return r.AllocateRegion()
	case "Schemas":
		return r.AllocateSchemas()
	}
	return nil
}

func (r *TypeDeclRegion) MarshalBinary() ([]byte, error) {
	s := runtime.MakeSerializer()
	var err error
	err = s.WriteCount(len(r.FieldPool))
	if err != nil {
		return nil, err
	}
	err = s.WriteCount(len(r.StructPool))
	if err != nil {
		return nil, err
	}
	err = s.WriteCount(len(r.RegionPool))
	if err != nil {
		return nil, err
	}
	err = s.WriteCount(len(r.SchemasPool))
	if err != nil {
		return nil, err
	}
	for _, o := range r.FieldPool {
		s.WriteString(o.Name)
		s.WriteString(o.Type)
	}
	for _, o := range r.StructPool {
		s.WriteString(o.Name)
		err = s.WriteCount(len(o.Fields))
		if err != nil {
			return nil, err
		}
		for _, o0 := range o.Fields {
			err = s.WriteIndex(o0.PoolIndex, len(r.FieldPool))
			if err != nil {
				return nil, err
			}
		}
	}
	for _, o := range r.RegionPool {
		s.WriteString(o.Name)
		err = s.WriteCount(len(o.Struct))
		if err != nil {
			return nil, err
		}
		for _, o0 := range o.Struct {
			err = s.WriteIndex(o0.PoolIndex, len(r.StructPool))
			if err != nil {
				return nil, err
			}
		}
	}
	for _, o := range r.SchemasPool {
		err = s.WriteCount(len(o.Region))
		if err != nil {
			return nil, err
		}
		for _, o0 := range o.Region {
			err = s.WriteIndex(o0.PoolIndex, len(r.RegionPool))
			if err != nil {
				return nil, err
			}
		}
	}
	return s.Data(), nil
}

func (r *TypeDeclRegion) UnmarshalBinary(data []byte) error {
	d := runtime.MakeDeserializer(data)
	var index int
	var err error
	index, err = d.ReadCount()
	if err != nil {
		return err
	}
	for i := 0; i < index; i++ {
		r.AllocateField()
	}
	index, err = d.ReadCount()
	if err != nil {
		return err
	}
	for i := 0; i < index; i++ {
		r.AllocateStruct()
	}
	index, err = d.ReadCount()
	if err != nil {
		return err
	}
	for i := 0; i < index; i++ {
		r.AllocateRegion()
	}
	index, err = d.ReadCount()
	if err != nil {
		return err
	}
	for i := 0; i < index; i++ {
		r.AllocateSchemas()
	}
	for _, o := range r.FieldPool {
		o.Name, err = d.ReadString()
		if err != nil {
			return err
		}
		o.Type, err = d.ReadString()
		if err != nil {
			return err
		}
	}
	for _, o := range r.StructPool {
		o.Name, err = d.ReadString()
		if err != nil {
			return err
		}
		index, err = d.ReadCount()
		if err != nil {
			return err
		}
		o.Fields = make([]*Field, index)
		for i0, _ := range o.Fields {
			index, err = d.ReadIndex(len(r.FieldPool))
			if err != nil {
				return err
			}
			o.Fields[i0] = r.FieldPool[index]
		}
	}
	for _, o := range r.RegionPool {
		o.Name, err = d.ReadString()
		if err != nil {
			return err
		}
		index, err = d.ReadCount()
		if err != nil {
			return err
		}
		o.Struct = make([]*Struct, index)
		for i0, _ := range o.Struct {
			index, err = d.ReadIndex(len(r.StructPool))
			if err != nil {
				return err
			}
			o.Struct[i0] = r.StructPool[index]
		}
	}
	for _, o := range r.SchemasPool {
		index, err = d.ReadCount()
		if err != nil {
			return err
		}
		o.Region = make([]*Region, index)
		for i0, _ := range o.Region {
			index, err = d.ReadIndex(len(r.RegionPool))
			if err != nil {
				return err
			}
			o.Region[i0] = r.RegionPool[index]
		}
	}
	return nil
}

func CreateTypeDeclRegion() *TypeDeclRegion {
	return &TypeDeclRegion{}
}

var typeDeclRegionSchema = &runtime.RegionSchema{Name: "TypeDecl", GoType: (*TypeDeclRegion)(nil)}

func init() {

	fieldSchema.Fields = []*runtime.FieldSchema{
		{Name: "name", Type: &runtime.StringSchema{}},
		{Name: "type", Type: &runtime.StringSchema{}},
	}

	structSchema.Fields = []*runtime.FieldSchema{
		{Name: "name", Type: &runtime.StringSchema{}},
		{Name: "fields", Type: fieldSchema.List()},
	}

	regionSchema.Fields = []*runtime.FieldSchema{
		{Name: "name", Type: &runtime.StringSchema{}},
		{Name: "struct", Type: structSchema.List()},
	}

	schemasSchema.Fields = []*runtime.FieldSchema{
		{Name: "region", Type: regionSchema.List()},
	}

	typeDeclRegionSchema.Structs = []*runtime.StructSchema{
		fieldSchema,
		structSchema,
		regionSchema,
		schemasSchema,
	}
	typeDeclRegionSchema.Init()
}
