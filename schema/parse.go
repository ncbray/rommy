// Package schema handles schema declarations.
package schema

import (
	"github.com/ncbray/rommy/human"
)

//go:generate rommyc schema.rommy --go_out .

func ParseSchema(file string, data []byte) (*TypeDeclRegion, *Schemas, bool) {
	region := CreateTypeDeclRegion()

	generic_result, ok := human.ParseFile(file, data, region)
	if !ok {
		return nil, nil, false
	}
	result, ok := generic_result.(*Schemas)
	if !ok {
		return nil, nil, false
	}
	return region, result, true
}
