package runtime

type Region interface {
	Schema() *RegionSchema
	Allocate(name string) interface{}
}

type Struct interface {
	Schema() *StructSchema
}
