// Command rommygen generates Go sources from schema declarations.
package main

import (
	"fmt"
	"github.com/ncbray/cmdline"
	"github.com/ncbray/compilerutil/fs"
	"github.com/ncbray/compilerutil/names"
	"github.com/ncbray/compilerutil/writer"
	"github.com/ncbray/rommy/runtime"
	"github.com/ncbray/rommy/schema"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func goTypeRef(t runtime.TypeSchema) string {
	switch t := t.(type) {
	case *runtime.IntegerSchema:
		// HACK canonical name matches Go type.
		return t.CanonicalName()
	case *runtime.StringSchema:
		return "string"
	case *runtime.StructSchema:
		return "*" + t.Name
	case *runtime.ListSchema:
		return "[]" + goTypeRef(t.Element)
	default:
		panic(t)
	}
}

func fieldName(f *runtime.FieldSchema) string {
	return names.JoinCamelCase(names.SplitSnakeCase(f.Name), true)
}

func structSchemaName(s *runtime.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Schema"), false)
}

func schemaFieldType(t runtime.TypeSchema) string {
	switch t := t.(type) {
	case *runtime.IntegerSchema:
		return fmt.Sprintf("&runtime.IntegerSchema{Bits: %d, Unsigned: %v}", t.Bits, t.Unsigned)
	case *runtime.StringSchema:
		return "&runtime.StringSchema{}"
	case *runtime.StructSchema:
		return structSchemaName(t)
	case *runtime.ListSchema:
		return schemaFieldType(t.Element) + ".List()"
	default:
		panic(t)
	}
}

func regionStructName(r *runtime.RegionSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(r.Name+"Region"), true)
}

func regionSchemaName(r *runtime.RegionSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(r.Name+"RegionSchema"), false)
}

func poolField(r *runtime.RegionSchema, s *runtime.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Pool"), true)
}

func regionClonerName(r *runtime.RegionSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(r.Name+"Cloner"), true)
}

func mapingField(r *runtime.RegionSchema, s *runtime.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Map"), false)
}

func abortSerializeOnError(out *writer.TabbedWriter) {
	out.WriteLine("if err != nil {")
	out.Indent()
	out.WriteLine("return nil, err")
	out.Dedent()
	out.WriteLine("}")
}

func abortDeserializeOnError(out *writer.TabbedWriter) {
	out.WriteLine("if err != nil {")
	out.Indent()
	out.WriteLine("return err")
	out.Dedent()
	out.WriteLine("}")
}

func serialize(path string, level int, r *runtime.RegionSchema, t runtime.TypeSchema, out *writer.TabbedWriter) {
	switch t := t.(type) {
	case *runtime.IntegerSchema:
		out.WriteString("s.Write")
		out.WriteString(names.Capitalize(t.CanonicalName()))
		out.WriteString("(")
		out.WriteString(path)
		out.WriteString(")")
		out.EndOfLine()
	case *runtime.StringSchema:
		out.WriteString("s.WriteString(")
		out.WriteString(path)
		out.WriteString(")")
		out.EndOfLine()
	case *runtime.StructSchema:
		out.WriteString("err = s.WriteIndex(")
		out.WriteString(path)
		out.WriteString(".PoolIndex, len(r.")
		out.WriteString(poolField(r, t))
		out.WriteString("))")
		out.EndOfLine()
		abortSerializeOnError(out)
	case *runtime.ListSchema:
		out.WriteString("err = s.WriteCount(len(")
		out.WriteString(path)
		out.WriteString("))")
		out.EndOfLine()
		abortSerializeOnError(out)

		child_path := "o" + strconv.Itoa(level)
		out.WriteString("for _, ")
		out.WriteString(child_path)
		out.WriteString(" := range ")
		out.WriteString(path)
		out.WriteString(" {")
		out.EndOfLine()
		out.Indent()
		serialize(child_path, level+1, r, t.Element, out)
		out.Dedent()
		out.WriteLine("}")
	default:
		panic(t)
	}
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

func generateStructDecls(r *runtime.RegionSchema, s *runtime.StructSchema, out *writer.TabbedWriter) {
	out.EndOfLine()
	out.WriteString("type ")
	out.WriteString(s.Name)
	out.WriteString(" struct {")
	out.EndOfLine()
	out.Indent()
	out.WriteLine("PoolIndex int")
	for _, f := range s.Fields {
		out.WriteString(fieldName(f))
		out.WriteString(" ")
		out.WriteString(goTypeRef(f.Type))
		out.EndOfLine()
	}
	out.Dedent()
	out.WriteLine("}")

	// Global variable holding the schema.
	schemaName := structSchemaName(s)

	out.EndOfLine()
	out.WriteString("func (s *")
	out.WriteString(s.Name)
	out.WriteString(") Schema() *runtime.StructSchema {")
	out.EndOfLine()
	out.Indent()
	out.WriteString("return ")
	out.WriteString(schemaName)
	out.EndOfLine()
	out.Dedent()
	out.WriteLine("}")

	out.EndOfLine()
	out.WriteString("var ")
	out.WriteString(schemaName)
	out.WriteString(" = &runtime.StructSchema{")
	out.WriteString("Name: ")
	out.WriteString(strconv.Quote(s.Name))
	out.WriteString(", GoType: (*")
	out.WriteString(s.Name)
	out.WriteString(")(nil)}")
	out.EndOfLine()
}

func generateRegionSerialization(r *runtime.RegionSchema, out *writer.TabbedWriter) {
	structName := regionStructName(r)

	// Serializer
	out.EndOfLine()
	out.WriteString("func (r *")
	out.WriteString(structName)
	out.WriteString(") MarshalBinary() ([]byte, error) {")
	out.EndOfLine()
	out.Indent()
	out.WriteLine("s := runtime.MakeSerializer()")
	out.WriteLine("var err error")

	// indexs
	for _, s := range r.Structs {
		f := poolField(r, s)
		out.WriteString("err = s.WriteCount(len(r.")
		out.WriteString(f)
		out.WriteString("))")
		out.EndOfLine()
		abortSerializeOnError(out)
	}
	// Values
	for _, s := range r.Structs {
		f := poolField(r, s)
		out.WriteString("for _, o := range r.")
		out.WriteString(f)
		out.WriteString(" {")
		out.EndOfLine()
		out.Indent()

		for _, f := range s.Fields {
			path := "o." + fieldName(f)
			serialize(path, 0, r, f.Type, out)
		}
		out.Dedent()
		out.WriteLine("}")
	}
	out.WriteLine("return s.Data(), nil")
	out.Dedent()
	out.WriteLine("}")

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

func generateValueClone(src_path string, dst_path string, level int, t runtime.TypeSchema, r *runtime.RegionSchema, out *writer.TabbedWriter) {
	switch t := t.(type) {
	case *runtime.IntegerSchema, *runtime.StringSchema:
		out.WriteString(dst_path)
		out.WriteString(" = ")
		out.WriteString(src_path)
		out.EndOfLine()
	case *runtime.StructSchema:
		out.WriteString(dst_path)
		out.WriteString(" = c.Clone")
		out.WriteString(t.Name)
		out.WriteString("(")
		out.WriteString(src_path)
		out.WriteString(")")
		out.EndOfLine()
	case *runtime.ListSchema:
		out.WriteString(dst_path)
		out.WriteString(" = make(")
		out.WriteString(goTypeRef(t))
		out.WriteString(", len(")
		out.WriteString(src_path)
		out.WriteString("))")
		out.EndOfLine()

		child_index := "i" + strconv.Itoa(level)
		out.WriteLine("for " + child_index + ", _ := range " + src_path + " {")
		out.Indent()
		index_op := "[" + child_index + "]"
		generateValueClone(src_path+index_op, dst_path+index_op, level+1, t.Element, r, out)
		out.Dedent()
		out.WriteLine("}")
		// Copy
	default:
		panic(t)
	}
}

func generateRegionCloner(r *runtime.RegionSchema, out *writer.TabbedWriter) {
	structName := regionStructName(r)
	clonerName := regionClonerName(r)

	out.EndOfLine()
	out.WriteString("type ")
	out.WriteString(clonerName)
	out.WriteString(" struct {")
	out.EndOfLine()
	out.Indent()

	out.WriteString("src *")
	out.WriteString(structName)
	out.EndOfLine()

	out.WriteString("dst *")
	out.WriteString(structName)
	out.EndOfLine()

	for _, s := range r.Structs {
		f := mapingField(r, s)
		out.WriteString(f)
		out.WriteString(" []*")
		out.WriteString(s.Name)
		out.EndOfLine()
	}

	out.Dedent()
	out.WriteLine("}")

	// Constructor
	out.EndOfLine()
	out.WriteString("func Create")
	out.WriteString(clonerName)
	out.WriteString("(src *")
	out.WriteString(structName)
	out.WriteString(", dst *")
	out.WriteString(structName)
	out.WriteString(") *")
	out.WriteString(clonerName)
	out.WriteString(" {")
	out.EndOfLine()
	out.Indent()
	out.WriteString("c := &")
	out.WriteString(clonerName)
	out.WriteString("{")
	out.EndOfLine()
	out.Indent()
	out.WriteLine("src: src,")
	out.WriteLine("dst: dst,")
	for _, s := range r.Structs {
		f := mapingField(r, s)
		out.WriteString(f)
		out.WriteString(": make([]*")
		out.WriteString(s.Name)
		out.WriteString(", len(src.")
		out.WriteString(poolField(r, s))
		out.WriteString(")),")
		out.EndOfLine()
	}
	out.Dedent()
	out.WriteLine("}")
	out.WriteLine("return c")
	out.Dedent()
	out.WriteLine("}")

	// Struct clone methods.
	for _, s := range r.Structs {
		out.EndOfLine()
		out.WriteString("func (c *")
		out.WriteString(clonerName)
		out.WriteString(") Clone")
		out.WriteString(s.Name)
		out.WriteString("(src *")
		out.WriteString(s.Name)
		out.WriteString(") *")
		out.WriteString(s.Name)
		out.WriteString(" {")
		out.EndOfLine()
		out.Indent()

		f := mapingField(r, s)
		out.WriteString("dst := c.")
		out.WriteString(f)
		out.WriteString("[src.PoolIndex]")
		out.EndOfLine()

		// Early out
		out.WriteLine("if dst != nil {")
		out.Indent()
		out.WriteLine("return dst")
		out.Dedent()
		out.WriteLine("}")

		out.WriteString("dst = c.dst.Allocate")
		out.WriteString(s.Name)
		out.WriteString("()")
		out.EndOfLine()

		out.WriteString("c.")
		out.WriteString(f)
		out.WriteString("[src.PoolIndex] = dst")
		out.EndOfLine()

		// Deep clone
		for _, f := range s.Fields {
			fn := fieldName(f)
			generateValueClone("src."+fn, "dst."+fn, 0, f.Type, r, out)
		}

		out.WriteLine("return dst")
		out.Dedent()
		out.WriteLine("}")
	}
}

func generateRegionDecls(r *runtime.RegionSchema, out *writer.TabbedWriter) {
	for _, s := range r.Structs {
		generateStructDecls(r, s, out)
	}

	structName := regionStructName(r)
	schemaName := regionSchemaName(r)

	out.EndOfLine()
	out.WriteString("type ")
	out.WriteString(structName)
	out.WriteString(" struct {")
	out.EndOfLine()
	out.Indent()
	for _, s := range r.Structs {
		out.WriteString(poolField(r, s))
		out.WriteString(" []*")
		out.WriteString(s.Name)
		out.EndOfLine()
	}
	out.Dedent()
	out.WriteLine("}")

	// Schema getter.
	out.EndOfLine()
	out.WriteString("func (r *")
	out.WriteString(structName)
	out.WriteString(") Schema() *runtime.RegionSchema {")
	out.EndOfLine()
	out.Indent()
	out.WriteString("return ")
	out.WriteString(schemaName)
	out.EndOfLine()
	out.Dedent()
	out.WriteLine("}")

	// Concrete child allocators.
	for _, s := range r.Structs {
		out.EndOfLine()
		out.WriteString("func (r *")
		out.WriteString(structName)
		out.WriteString(") Allocate")
		out.WriteString(s.Name)
		out.WriteString("() *")
		out.WriteString(s.Name)
		out.WriteString(" {")
		out.EndOfLine()
		out.Indent()
		out.WriteString("o := &")
		out.WriteString(s.Name)
		out.WriteString("{}")
		out.EndOfLine()

		f := poolField(r, s)

		out.WriteString("o.PoolIndex = len(r.")
		out.WriteString(f)
		out.WriteString(")")
		out.EndOfLine()

		out.WriteString("r.")
		out.WriteString(f)
		out.WriteString(" = append(r.")
		out.WriteString(f)
		out.WriteString(", o)")
		out.EndOfLine()

		out.WriteLine("return o")
		out.Dedent()
		out.WriteLine("}")
	}

	// Generic child allocator
	out.EndOfLine()
	out.WriteString("func (r *")
	out.WriteString(structName)
	out.WriteString(") Allocate(name string) interface{} {")
	out.EndOfLine()
	out.Indent()
	out.WriteLine("switch name {")
	for _, s := range r.Structs {
		out.WriteString("case ")
		out.WriteString(strconv.Quote(s.Name))
		out.WriteString(":")
		out.EndOfLine()
		out.Indent()
		out.WriteString("return r.Allocate")
		out.WriteString(s.Name)
		out.WriteString("()")
		out.EndOfLine()
		out.Dedent()
	}
	out.WriteLine("}")
	out.WriteLine("return nil")
	out.Dedent()
	out.WriteLine("}")

	generateRegionSerialization(r, out)

	// Constructor
	out.EndOfLine()
	out.WriteString("func Create")
	out.WriteString(structName)
	out.WriteString("() *")
	out.WriteString(structName)
	out.WriteString(" {")
	out.EndOfLine()
	out.Indent()
	out.WriteString("return &")
	out.WriteString(structName)
	out.WriteString("{}")
	out.EndOfLine()
	out.Dedent()
	out.WriteLine("}")

	out.EndOfLine()
	out.WriteString("var ")
	out.WriteString(schemaName)
	out.WriteString(" = &runtime.RegionSchema{")
	out.WriteString("Name: ")
	out.WriteString(strconv.Quote(r.Name))
	out.WriteString(", GoType: (*")
	out.WriteString(structName)
	out.WriteString(")(nil)}")
	out.EndOfLine()

	generateRegionCloner(r, out)
}

func generateStructInit(r *runtime.RegionSchema, s *runtime.StructSchema, out *writer.TabbedWriter) {
	schemaName := structSchemaName(s)

	out.EndOfLine()
	out.WriteString(schemaName)
	out.WriteString(".Fields = []*runtime.FieldSchema{")
	out.EndOfLine()

	out.Indent()
	for _, f := range s.Fields {
		out.WriteString("{Name: ")
		out.WriteString(strconv.Quote(f.Name))
		out.WriteString(", Type: ")
		out.WriteString(schemaFieldType(f.Type))
		out.WriteString("},")
		out.EndOfLine()
	}
	out.Dedent()
	out.WriteLine("}")
}

func generateRegionInit(r *runtime.RegionSchema, out *writer.TabbedWriter) {
	for _, s := range r.Structs {
		generateStructInit(r, s, out)
	}

	schemaName := regionSchemaName(r)

	out.EndOfLine()
	out.WriteString(schemaName)
	out.WriteString(".Structs = []*runtime.StructSchema{")
	out.EndOfLine()

	out.Indent()
	for _, s := range r.Structs {
		out.WriteString(structSchemaName(s))
		out.WriteString(",")
		out.EndOfLine()
	}
	out.Dedent()
	out.WriteLine("}")

	out.WriteString(schemaName)
	out.WriteString(".Init()")
	out.EndOfLine()

}

func generateGoSrc(pkg string, regions []*runtime.RegionSchema, out *writer.TabbedWriter) {
	// Header
	out.WriteString("package ")
	out.WriteString(pkg)
	out.EndOfLine()
	out.EndOfLine()
	out.WriteLine("/* Generated with rommygen, do not edit by hand. */")
	out.EndOfLine()

	out.WriteLine("import (")
	out.Indent()
	out.WriteLine("\"github.com/ncbray/rommy/runtime\"")
	out.Dedent()
	out.WriteLine(")")

	for _, r := range regions {
		generateRegionDecls(r, out)
	}

	// Init
	out.EndOfLine()
	out.WriteLine("func init() {")
	out.Indent()

	for _, r := range regions {
		generateRegionInit(r, out)
	}

	out.Dedent()
	out.WriteLine("}")

}

func formatGoFile(src fs.DataInput, dst fs.DataOutput) error {
	data, err := src.GetBytes()
	if err != nil {
		return err
	}
	data, err = format.Source(data)
	if err != nil {
		return err
	}
	return dst.SetBytes(data)
}

func main() {
	inputFile := &cmdline.FilePath{
		MustExist: true,
	}

	var input string

	app := cmdline.MakeApp("rommygen")
	app.RequiredArgs([]*cmdline.Argument{
		{
			Name:  "input",
			Value: inputFile.Set(&input),
		},
	})
	app.Run(os.Args[1:])

	data, err := ioutil.ReadFile(input)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	dir, file := filepath.Split(input)
	ext := filepath.Ext(file)
	base := file[0 : len(file)-len(ext)]

	// Infer the go package from the absolute path.
	abs_input, err := filepath.Abs(input)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	abs_dir, _ := filepath.Split(abs_input)
	pkg := filepath.Base(abs_dir)
	if pkg == "" || pkg == "." {
		println("Cannot infer package for file " + abs_input)
		os.Exit(1)
	}

	_, result, ok := schema.ParseSchema(input, data)
	if !ok {
		os.Exit(1)
	}
	//runtime.DumpText(result, os.Stdout)
	regions := schema.Resolve(result)

	tmp, err := fs.MakeTempDir("rommygen_")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer tmp.Cleanup()
	buffered := fs.MakeBufferedFileSystem(tmp)

	tmpf := buffered.TempFile()
	ow, err := tmpf.GetWriter()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	out := writer.MakeTabbedWriter("\t", ow)
	generateGoSrc(pkg, regions, out)

	outf := buffered.OutputFile(filepath.Join(dir, base+".go"), 0644)
	err = formatGoFile(tmpf, outf)
	if err != nil {
		println("formatting error - " + err.Error())
		os.Exit(1)
	}

	buffered.Commit()
}
