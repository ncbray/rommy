package rommy

import (
	"github.com/ncbray/compilerutil/names"
)

type Namespace struct {
	types map[string]TypeSchema
}

func (ns *Namespace) Get(name string) (TypeSchema, bool) {
	t, ok := ns.types[name]
	return t, ok
}

func (ns *Namespace) set(name string, t TypeSchema) {
	if ns.types == nil {
		ns.types = map[string]TypeSchema{}
	}
	ns.types[name] = t
}

func (ns *Namespace) Register(t TypeSchema) {
	var name string
	switch t := t.(type) {
	case *StructSchema:
		t.Init()
		name = t.Name
	default:
		panic(t)
	}
	ns.set(name, t)
}

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
	GoType    interface{}
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
