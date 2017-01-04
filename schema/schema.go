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
