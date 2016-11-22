package schema

import (
	"github.com/ncbray/rommy"
)

func ParseSchema(file string, data []byte) (*Schemas, bool) {
	sources := rommy.CreateSourceSet()
	status := &rommy.Status{Sources: sources}

	info := sources.Add(file, data)
	e := rommy.ParseData(info, data, status)
	if status.ShouldStop() {
		return nil, false
	}
	result, ok := rommy.HandleData(Namespace, e, nil, status)
	if !ok {
		return nil, false
	}
	schemas, ok := result.(*Schemas)
	// TODO error handling?
	return schemas, ok
}
