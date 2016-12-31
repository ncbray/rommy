package rommy

import (
	"github.com/ncbray/compilerutil/writer"
	"io"
	"reflect"
	"strconv"
)

type RommyStruct interface {
	Schema() *StructSchema
}

func isDefaultValue(o reflect.Value, schema TypeSchema) bool {
	switch schema := schema.(type) {
	case *StringSchema:
		return o.String() == ""
	case *IntegerSchema:
		return o.Int() == 0
	case *StructSchema:
		return false
	case *ListSchema:
		return o.Len() == 0
	default:
		panic(schema)
	}
}

func dumpStruct(o reflect.Value, schema TypeSchema, expected TypeSchema, out *writer.TabbedWriter) {
	switch schema := schema.(type) {
	case *StringSchema:
		// TODO custom string quoting.
		out.WriteString(strconv.Quote(o.String()))
	case *IntegerSchema:
		out.WriteString(strconv.FormatInt(o.Int(), 10))
	case *StructSchema:
		o = o.Elem()
		if schema != expected {
			out.WriteString(schema.Name)
			out.WriteString(" ")
		}
		out.WriteString("{")
		out.EndOfLine()
		out.Indent()
		for _, f := range schema.Fields {
			child := o.FieldByName(f.GoName())
			if isDefaultValue(child, f.Type) {
				continue
			}
			out.WriteString(f.Name)
			out.WriteString(": ")
			dumpStruct(child, f.Type, f.Type, out)
			out.WriteString(",")
			out.EndOfLine()
		}
		out.Dedent()
		out.WriteString("}")
	case *ListSchema:
		out.WriteString("[")
		out.EndOfLine()
		out.Indent()
		for i := 0; i < o.Len(); i++ {
			child := o.Index(i)
			dumpStruct(child, schema.Element, schema.Element, out)
			out.WriteString(",")
			out.EndOfLine()
		}
		out.Dedent()
		out.WriteString("]")
	default:
		panic(schema)
	}
}

func DumpText(s RommyStruct, w io.Writer) {
	out := writer.MakeTabbedWriter("  ", w)
	dumpStruct(reflect.ValueOf(s), s.Schema(), nil, out)
	out.EndOfLine()
}
