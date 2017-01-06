package golang

import (
	"github.com/ncbray/compilerutil/names"
	"github.com/ncbray/compilerutil/writer"
	"github.com/ncbray/rommy/runtime"
	"strconv"
)

func abortDeserializeOnError(out *writer.TabbedWriter) {
	out.WriteLine("if err != nil {")
	out.Indent()
	out.WriteLine("return err")
	out.Dedent()
	out.WriteLine("}")
}

func deserialize(path string, level int, r *runtime.RegionSchema, t runtime.TypeSchema, out *writer.TabbedWriter) {
	switch t := t.(type) {
	case *runtime.IntegerSchema:
		out.WriteString(path)
		out.WriteString(", err = d.Read")
		out.WriteString(names.Capitalize(t.CanonicalName()))
		out.WriteString("()")
		out.EndOfLine()
		abortDeserializeOnError(out)
	case *runtime.StringSchema:
		out.WriteString(path)
		out.WriteString(", err = d.ReadString()")
		out.EndOfLine()
		abortDeserializeOnError(out)
	case *runtime.StructSchema:
		f := poolField(r, t)
		out.WriteString("index, err = d.ReadIndex(len(r.")
		out.WriteString(f)
		out.WriteString("))")
		out.EndOfLine()
		abortDeserializeOnError(out)

		out.WriteString(path)
		out.WriteString(" = r.")
		out.WriteString(f)
		out.WriteString("[index]")
		out.EndOfLine()
	case *runtime.ListSchema:
		out.WriteLine("index, err = d.ReadCount()")
		abortDeserializeOnError(out)
		out.WriteString(path)
		out.WriteString(" = make(")
		out.WriteString(goTypeRef(t))
		out.WriteString(", index)")
		out.EndOfLine()

		child_index := "i" + strconv.Itoa(level)
		out.WriteString("for ")
		out.WriteString(child_index)
		out.WriteString(", _ := range ")
		out.WriteString(path)
		out.WriteString(" {")
		out.EndOfLine()
		out.Indent()
		deserialize(path+"["+child_index+"]", level+1, r, t.Element, out)
		out.Dedent()
		out.WriteLine("}")
	default:
		panic(t)
	}
}

func generateRegionDeserialize(r *runtime.RegionSchema, out *writer.TabbedWriter) {
	structName := regionStructName(r)

	// Deserializer
	out.EndOfLine()
	out.WriteString("func (r *")
	out.WriteString(structName)
	out.WriteString(") UnmarshalBinary(data []byte) error {")
	out.EndOfLine()
	out.Indent()
	out.WriteLine("d := runtime.MakeDeserializer(data)")
	out.WriteLine("var index int")
	out.WriteLine("var err error")

	// Allocate objects
	for _, s := range r.Structs {
		out.WriteLine("index, err = d.ReadCount()")
		abortDeserializeOnError(out)
		// TODO allocate exact count.
		out.WriteLine("for i := 0; i < index; i++ {")
		out.Indent()
		out.WriteString("r.Allocate")
		out.WriteString(s.Name)
		out.WriteString("()")
		out.EndOfLine()
		out.Dedent()
		out.WriteLine("}")
	}
	// Deserialize objects
	for _, s := range r.Structs {
		f := poolField(r, s)
		out.WriteString("for _, o := range r.")
		out.WriteString(f)
		out.WriteString(" {")
		out.EndOfLine()
		out.Indent()

		for _, f := range s.Fields {
			path := "o." + fieldName(f)
			deserialize(path, 0, r, f.Type, out)
		}
		out.Dedent()
		out.WriteLine("}")
	}
	out.WriteLine("return nil")
	out.Dedent()
	out.WriteLine("}")
}
