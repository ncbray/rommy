package main

import (
	"github.com/ncbray/cmdline"
	"github.com/ncbray/compilerutil/fs"
	"github.com/ncbray/compilerutil/names"
	"github.com/ncbray/compilerutil/writer"
	"github.com/ncbray/rommy"
	"github.com/ncbray/rommy/schema"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func goTypeRef(t rommy.TypeSchema) string {
	switch t := t.(type) {
	case *rommy.IntegerSchema:
		return "int32"
	case *rommy.StringSchema:
		return "string"
	case *rommy.StructSchema:
		return "*" + t.Name
	case *rommy.ListSchema:
		return "[]" + goTypeRef(t.Element)
	default:
		panic(t)
	}
}

func structSchemaName(s *rommy.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Schema"), false)
}

func schemaFieldType(t rommy.TypeSchema) string {
	switch t := t.(type) {
	case *rommy.IntegerSchema:
		return "&rommy.IntegerSchema{}"
	case *rommy.StringSchema:
		return "&rommy.StringSchema{}"
	case *rommy.StructSchema:
		return structSchemaName(t)
	case *rommy.ListSchema:
		return schemaFieldType(t.Element) + ".List()"
	default:
		panic(t)
	}
}

func regionStructName(r *rommy.RegionSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(r.Name+"Region"), true)
}

func regionSchemaName(r *rommy.RegionSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(r.Name+"RegionSchema"), false)
}

func poolField(r *rommy.RegionSchema, s *rommy.StructSchema) string {
	return names.JoinCamelCase(names.SplitCamelCase(s.Name+"Pool"), true)
}

func generateStructDecls(r *rommy.RegionSchema, s *rommy.StructSchema, out *writer.TabbedWriter) {
	out.EndOfLine()
	out.WriteString("type ")
	out.WriteString(s.Name)
	out.WriteString(" struct {")
	out.EndOfLine()
	out.Indent()
	for _, f := range s.Fields {
		out.WriteString(names.JoinCamelCase(names.SplitSnakeCase(f.Name), true))
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
	out.WriteString(") Schema() *rommy.StructSchema {")
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
	out.WriteString(" = &rommy.StructSchema{")
	out.WriteString("Name: ")
	out.WriteString(strconv.Quote(s.Name))
	out.WriteString(", GoType: (*")
	out.WriteString(s.Name)
	out.WriteString(")(nil)}")
	out.EndOfLine()
}

func generateRegionDecls(r *rommy.RegionSchema, out *writer.TabbedWriter) {
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
		out.WriteString("[]*")
		out.WriteString(s.Name)
		out.EndOfLine()
	}
	out.Dedent()
	out.WriteLine("}")

	out.EndOfLine()
	out.WriteString("func (r *")
	out.WriteString(structName)
	out.WriteString(") Schema() *rommy.RegionSchema {")
	out.EndOfLine()
	out.Indent()
	out.WriteString("return ")
	out.WriteString(schemaName)
	out.EndOfLine()
	out.Dedent()
	out.WriteLine("}")

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
	out.WriteString(" = &rommy.RegionSchema{")
	out.WriteString("Name: ")
	out.WriteString(strconv.Quote(r.Name))
	out.WriteString(", GoType: (*")
	out.WriteString(structName)
	out.WriteString(")(nil)}")
	out.EndOfLine()
}

func generateStructInit(r *rommy.RegionSchema, s *rommy.StructSchema, out *writer.TabbedWriter) {
	schemaName := structSchemaName(s)

	out.EndOfLine()
	out.WriteString(schemaName)
	out.WriteString(".Fields = []*rommy.FieldSchema{")
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

func generateRegionInit(r *rommy.RegionSchema, out *writer.TabbedWriter) {
	for _, s := range r.Structs {
		generateStructInit(r, s, out)
	}

	schemaName := regionSchemaName(r)

	out.EndOfLine()
	out.WriteString(schemaName)
	out.WriteString(".Structs = []*rommy.StructSchema{")
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

func generateGoSrc(pkg string, regions []*rommy.RegionSchema, out *writer.TabbedWriter) {
	// Header
	out.WriteString("package ")
	out.WriteString(pkg)
	out.EndOfLine()
	out.EndOfLine()
	out.WriteLine("/* Generated with rommygen, do not edit by hand. */")
	out.EndOfLine()

	out.WriteLine("import (")
	out.Indent()
	out.WriteLine("\"github.com/ncbray/rommy\"")
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
	//rommy.DumpText(result, os.Stdout)
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
		println(err.Error())
		os.Exit(1)
	}

	buffered.Commit()
}
