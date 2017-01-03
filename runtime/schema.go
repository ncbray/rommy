package runtime

import (
	"github.com/ncbray/compilerutil/names"
)

type TypeSchema interface {
	List() *ListSchema
	CanHold(other TypeSchema) bool
	CanonicalName() string
}

type StringSchema struct {
	listCache *ListSchema
}

func (s *StringSchema) CanHold(other TypeSchema) bool {
	_, ok := other.(*StringSchema)
	return ok
}

func (s *StringSchema) List() *ListSchema {
	if s.listCache == nil {
		s.listCache = &ListSchema{Element: s}
	}
	return s.listCache
}

func (s *StringSchema) CanonicalName() string {
	return "string"
}

type IntegerSchema struct {
	listCache *ListSchema
}

func (s *IntegerSchema) List() *ListSchema {
	if s.listCache == nil {
		s.listCache = &ListSchema{Element: s}
	}
	return s.listCache
}

func (s *IntegerSchema) CanHold(other TypeSchema) bool {
	_, ok := other.(*IntegerSchema)
	return ok
}

func (s *IntegerSchema) CanonicalName() string {
	return "int"
}

type FieldSchema struct {
	Name string
	Type TypeSchema
	ID   int
}

func (f *FieldSchema) GoName() string {
	return names.JoinCamelCase(names.SplitSnakeCase(f.Name), true)
}

type ListSchema struct {
	Element   TypeSchema
	listCache *ListSchema
}

func (s *ListSchema) List() *ListSchema {
	if s.listCache == nil {
		s.listCache = &ListSchema{Element: s}
	}
	return s.listCache
}

func (s *ListSchema) CanHold(other TypeSchema) bool {
	return s == other
}

func (s *ListSchema) CanonicalName() string {
	return "[]" + s.Element.CanonicalName()
}

type StructSchema struct {
	Name      string
	Fields    []*FieldSchema
	FieldLUT  map[string]*FieldSchema
	listCache *ListSchema
	GoType    Struct
}

func (s *StructSchema) Init() *StructSchema {
	s.FieldLUT = map[string]*FieldSchema{}
	for i, f := range s.Fields {
		f.ID = i
		s.FieldLUT[f.Name] = f
	}
	return s
}

func (s *StructSchema) List() *ListSchema {
	if s.listCache == nil {
		s.listCache = &ListSchema{Element: s}
	}
	return s.listCache
}

func (s *StructSchema) CanHold(other TypeSchema) bool {
	return s == other
}

func (s *StructSchema) CanonicalName() string {
	return s.Name
}

type RegionSchema struct {
	Name      string
	Structs   []*StructSchema
	StructLUT map[string]*StructSchema
	GoType    Region
}

func (r *RegionSchema) Init() *RegionSchema {
	r.StructLUT = map[string]*StructSchema{}
	for _, s := range r.Structs {
		s.Init()
		r.StructLUT[s.Name] = s
	}
	return r
}
