// Command rommyc generates Go sources from schema declarations.
package main

import (
	"github.com/ncbray/cmdline"
	"github.com/ncbray/compilerutil/fs"
	"github.com/ncbray/compilerutil/writer"
	"github.com/ncbray/rommy/generate/golang"
	"github.com/ncbray/rommy/generate/haxe"
	"github.com/ncbray/rommy/runtime"
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

func baseName(path string) string {
	_, file := filepath.Split(path)
	ext := filepath.Ext(file)
	return file[0 : len(file)-len(ext)]
}

func generateGo(input_file string, regions []*runtime.RegionSchema, output_dir string, buffered fs.BufferedFileSystem) {
	// Infer the go package from the absolute path of the output directory.
	abs_output_dir, err := filepath.Abs(output_dir)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	pkg := filepath.Base(abs_output_dir)
	if pkg == "" || pkg == "." {
		println("Cannot infer package for " + abs_output_dir)
		os.Exit(1)
	}

	tmpf := buffered.TempFile()
	ow, err := tmpf.GetWriter()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	out := writer.MakeTabbedWriter("\t", ow)
	golang.GenerateSource(pkg, regions, out)

	outf := buffered.OutputFile(filepath.Join(output_dir, baseName(input_file)+".go"), 0644)
	err = formatGoFile(tmpf, outf)
	if err != nil {
		println("formatting error - " + err.Error())
		os.Exit(1)
	}
}

func main() {
	inputFile := &cmdline.FilePath{
		MustExist: true,
	}
	outputFile := &cmdline.FilePath{
		MustExist: false,
	}

	var input string
	var go_out string
	var haxe_out string
	var haxe_package string

	app := cmdline.MakeApp("rommyc")
	app.Flags([]*cmdline.Flag{
		{
			Long:  "go_out",
			Value: outputFile.Set(&go_out),
		},
		{
			Long:  "haxe_out",
			Value: outputFile.Set(&haxe_out),
		},
		{
			Long:  "haxe_package",
			Value: cmdline.String.Set(&haxe_package),
		},
	})
	app.RequiredArgs([]*cmdline.Argument{
		{
			Name:  "input",
			Value: inputFile.Set(&input),
		},
	})
	app.Run(os.Args[1:])

	if go_out == "" && haxe_out == "" {
		println("ERROR no outputs specified for " + input)
		os.Exit(1)
	}

	if haxe_out != "" {
		if haxe_package == "" {
			println("ERROR haxe package not specified")
			os.Exit(1)
		}
	} else {
		if haxe_package != "" {
			println("ERROR haxe package specified when not generating haxe")
			os.Exit(1)
		}
	}

	data, err := ioutil.ReadFile(input)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	_, result, ok := schema.ParseSchema(input, data)
	if !ok {
		os.Exit(1)
	}
	//runtime.DumpText(result, os.Stdout)
	regions := schema.Resolve(result)

	tmp, err := fs.MakeTempDir("rommyc_")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer tmp.Cleanup()
	buffered := fs.MakeBufferedFileSystem(tmp)

	if go_out != "" {
		generateGo(input, regions, go_out, buffered)
	}

	if haxe_out != "" {
		err = haxe.GenerateSources(input, regions, haxe_out, haxe_package, buffered)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
	}

	buffered.Commit()
}
