// Command rommygen generates Go sources from schema declarations.
package main

import (
	"github.com/ncbray/cmdline"
	"github.com/ncbray/compilerutil/fs"
	"github.com/ncbray/compilerutil/writer"
	"github.com/ncbray/rommy/generate/golang"
	"github.com/ncbray/rommy/schema"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
)

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
	golang.GenerateSource(pkg, regions, out)

	outf := buffered.OutputFile(filepath.Join(dir, base+".go"), 0644)
	err = formatGoFile(tmpf, outf)
	if err != nil {
		println("formatting error - " + err.Error())
		os.Exit(1)
	}

	buffered.Commit()
}
