package schema

/* Generated with rommygen, do not edit by hand. */

import (
	"github.com/ncbray/rommy"
)

type Field struct {
	Name string
	Type string
}

func (s *Field) Schema() *rommy.StructSchema {
	return fieldSchema
}

var fieldSchema = &rommy.StructSchema{Name: "Field", GoType: (*Field)(nil)}

type Struct struct {
	Name   string
	Fields []*Field
}

func (s *Struct) Schema() *rommy.StructSchema {
	return structSchema
}

var structSchema = &rommy.StructSchema{Name: "Struct", GoType: (*Struct)(nil)}

type Region struct {
	Name   string
	Struct []*Struct
}

func (s *Region) Schema() *rommy.StructSchema {
	return regionSchema
}

var regionSchema = &rommy.StructSchema{Name: "Region", GoType: (*Region)(nil)}

type Schemas struct {
	Region []*Region
}

func (s *Schemas) Schema() *rommy.StructSchema {
	return schemasSchema
}

var schemasSchema = &rommy.StructSchema{Name: "Schemas", GoType: (*Schemas)(nil)}

type TypeDeclRegion struct {
	FieldPool   []*Field
	StructPool  []*Struct
	RegionPool  []*Region
	SchemasPool []*Schemas
}

func (r *TypeDeclRegion) Schema() *rommy.RegionSchema {
	return typeDeclRegionSchema
}

func (r *TypeDeclRegion) AllocateField() *Field {
	o := &Field{}
	r.FieldPool = append(r.FieldPool, o)
	return o
}

func (r *TypeDeclRegion) AllocateStruct() *Struct {
	o := &Struct{}
	r.StructPool = append(r.StructPool, o)
	return o
}

func (r *TypeDeclRegion) AllocateRegion() *Region {
	o := &Region{}
	r.RegionPool = append(r.RegionPool, o)
	return o
}

func (r *TypeDeclRegion) AllocateSchemas() *Schemas {
	o := &Schemas{}
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

var typeDeclRegionSchema = &rommy.RegionSchema{Name: "TypeDecl", GoType: (*TypeDeclRegion)(nil)}

func init() {

	fieldSchema.Fields = []*rommy.FieldSchema{
		{Name: "name", Type: &rommy.StringSchema{}},
		{Name: "type", Type: &rommy.StringSchema{}},
	}

	structSchema.Fields = []*rommy.FieldSchema{
		{Name: "name", Type: &rommy.StringSchema{}},
		{Name: "fields", Type: fieldSchema.List()},
	}

	regionSchema.Fields = []*rommy.FieldSchema{
		{Name: "name", Type: &rommy.StringSchema{}},
		{Name: "struct", Type: structSchema.List()},
	}

	schemasSchema.Fields = []*rommy.FieldSchema{
		{Name: "region", Type: regionSchema.List()},
	}

	typeDeclRegionSchema.Structs = []*rommy.StructSchema{
		fieldSchema,
		structSchema,
		regionSchema,
		schemasSchema,
	}
	typeDeclRegionSchema.Init()
}
