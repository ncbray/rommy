package schema

import (
	"github.com/ncbray/rommy"
)

//go:generate rommygen schema.rommy

func ParseSchema(file string, data []byte) (*TypeDeclRegion, *Schemas, bool) {
	region := CreateTypeDeclRegion()

	generic_result, ok := rommy.ParseFile(file, data, region)
	if !ok {
		return nil, nil, false
	}
	result, ok := generic_result.(*Schemas)
	if !ok {
		return nil, nil, false
	}
	return region, result, true
}
