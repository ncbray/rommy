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

func generateGoSrc(pkg string, structs []*rommy.StructSchema, out *writer.TabbedWriter) {
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

	// Type decls
	for _, s := range structs {
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

	// Init
	out.EndOfLine()
	out.WriteLine("var Namespace *rommy.Namespace")

	out.EndOfLine()
	out.WriteLine("func init() {")
	out.Indent()
	out.WriteLine("Namespace = &rommy.Namespace{}")

	for _, s := range structs {
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

		out.WriteString("Namespace.Register(")
		out.WriteString(schemaName)
		out.WriteString(")")
		out.EndOfLine()

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
	pkg := filepath.Base(dir)

	result, ok := schema.ParseSchema(input, data)
	if !ok {
		os.Exit(1)
	}
	//rommy.DumpText(result, os.Stdout)
	structs := schema.Resolve(result)

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
	generateGoSrc(pkg, structs, out)

	outf := buffered.OutputFile(filepath.Join(dir, base+".go"), 0644)
	err = formatGoFile(tmpf, outf)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	buffered.Commit()
}