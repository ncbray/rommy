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

type Schemas struct {
	Struct []*Struct
}

func (s *Schemas) Schema() *rommy.StructSchema {
	return schemasSchema
}

var schemasSchema = &rommy.StructSchema{Name: "Schemas", GoType: (*Schemas)(nil)}

var Namespace *rommy.Namespace

func init() {
	Namespace = &rommy.Namespace{}

	fieldSchema.Fields = []*rommy.FieldSchema{
		{Name: "name", Type: &rommy.StringSchema{}},
		{Name: "type", Type: &rommy.StringSchema{}},
	}
	Namespace.Register(fieldSchema)

	structSchema.Fields = []*rommy.FieldSchema{
		{Name: "name", Type: &rommy.StringSchema{}},
		{Name: "fields", Type: fieldSchema.List()},
	}
	Namespace.Register(structSchema)

	schemasSchema.Fields = []*rommy.FieldSchema{
		{Name: "struct", Type: structSchema.List()},
	}
	Namespace.Register(schemasSchema)
}
