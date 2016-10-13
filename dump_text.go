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

func dumpStruct(o reflect.Value, schema TypeSchema, out *writer.TabbedWriter) {
	switch schema := schema.(type) {
	case *StringSchema:
		// TODO custom string quoting.
		out.WriteString(strconv.Quote(o.String()))
	case *IntegerSchema:
		out.WriteString(strconv.FormatInt(o.Int(), 10))
	case *StructSchema:
		o = o.Elem()
		// TODO elide unneeded names.
		out.WriteString(schema.Name)
		out.WriteString(" {")
		out.EndOfLine()
		out.Indent()
		for _, f := range schema.Fields {
			// TODO elide zero values.
			out.WriteString(f.Name)
			out.WriteString(": ")
			child := o.FieldByName(f.GoName())
			dumpStruct(child, f.Type, out)
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
			dumpStruct(child, schema.Element, out)
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
	dumpStruct(reflect.ValueOf(s), s.Schema(), out)
	out.EndOfLine()
}
