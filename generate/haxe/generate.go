package haxe

import (
	"github.com/ncbray/compilerutil/fs"
	"github.com/ncbray/compilerutil/names"
	"github.com/ncbray/compilerutil/writer"
	"github.com/ncbray/rommy/runtime"
	"path/filepath"
	"strconv"
)

func generateStruct(pkg string, r *runtime.RegionSchema, s *runtime.StructSchema, out *writer.TabbedWriter) {
	out.WriteLine("package " + pkg + ";")

	out.EndOfLine()
	out.WriteLine("class " + structName(s) + " {")
	out.Indent()

	// Fields
	out.WriteLine("public var poolIndex:Int;")
	for _, f := range s.Fields {
		out.WriteLine("public var " + fieldName(f) + ":" + haxeTypeRef(f.Type) + ";")
	}

	// Constructor
	out.EndOfLine()
	out.WriteLine("@:allow(" + pkg + "." + regionName(r) + ")")
	out.WriteLine("private function new() {")
	out.Indent()
	out.Dedent()
	out.WriteLine("}")

	out.Dedent()
	out.WriteLine("}")
}

func abortDeserializeOnError(out *writer.TabbedWriter) {
	out.WriteLine("if (d.hasErrored()){")
	out.Indent()
	out.WriteLine("return false;")
	out.Dedent()
	out.WriteLine("}")
}

func deserialize(path string, level int, r *runtime.RegionSchema, t runtime.TypeSchema, out *writer.TabbedWriter) {
	switch t := t.(type) {
	case *runtime.IntegerSchema:
		out.WriteLine(path + " = d.read" + names.Capitalize(t.CanonicalName()) + "();")
		abortDeserializeOnError(out)
	case *runtime.StringSchema:
		out.WriteLine(path + " = d.readString();")
		abortDeserializeOnError(out)
	case *runtime.StructSchema:
		pf := poolField(r, t)
		out.WriteLine("index = d.readIndex(" + pf + ".length);")
		abortDeserializeOnError(out)
		out.WriteLine(path + " = " + pf + "[index];")
	case *runtime.ListSchema:
		out.WriteLine("index = d.readCount();")
		abortDeserializeOnError(out)
		child_index := "i" + strconv.Itoa(level)
		out.WriteLine(path + " = new " + haxeTypeRef(t) + "();")
		out.WriteLine("for (" + child_index + " in 0...index) {")
		out.Indent()
		// HACK
		out.WriteLine(path + ".push(null);")
		deserialize(path+"["+child_index+"]", level+1, r, t.Element, out)
		out.Dedent()
		out.WriteLine("}")
	default:
		panic(t)
	}
}

func generateRegion(pkg string, r *runtime.RegionSchema, out *writer.TabbedWriter) {
	out.WriteLine("package " + pkg + ";")

	out.EndOfLine()
	out.WriteLine("import haxe.io.Bytes;")
	out.WriteLine("import rommy.runtime.Deserializer;")

	out.EndOfLine()
	out.WriteLine("class " + regionName(r) + " {")
	out.Indent()

	// Fields
	for _, s := range r.Structs {
		out.WriteLine("public var " + poolField(r, s) + ":" + haxeTypeRef(s.List()) + ";")
	}

	// Constructor
	out.EndOfLine()
	out.WriteLine("public function new() {")
	out.Indent()
	for _, s := range r.Structs {
		out.WriteLine("this." + poolField(r, s) + " = [];")
	}
	out.Dedent()
	out.WriteLine("}")

	// Allocators
	for _, s := range r.Structs {
		pf := poolField(r, s)
		out.EndOfLine()
		out.WriteLine("public function allocate" + structName(s) + "():" + haxeTypeRef(s) + " {")
		out.Indent()
		out.WriteLine("var o = new " + haxeTypeRef(s) + "();")
		out.WriteLine("o.poolIndex = " + pf + ".length;")
		out.WriteLine(pf + ".push(o);")
		out.WriteLine("return o;")
		out.Dedent()
		out.WriteLine("}")
	}

	// Deserialize
	out.EndOfLine()
	out.WriteLine("public function deserialize(data:Bytes):Bool {")
	out.Indent()
	out.WriteLine("var d = new Deserializer(data);")
	out.WriteLine("var index:Int;")
	for _, s := range r.Structs {
		out.EndOfLine()
		out.WriteLine("index = d.readCount();")
		abortDeserializeOnError(out)
		out.WriteLine("for (i in 0...index) {")
		out.Indent()
		out.WriteLine("allocate" + structName(s) + "();")
		out.Dedent()
		out.WriteLine("}")
	}
	for _, s := range r.Structs {
		out.EndOfLine()
		out.WriteLine("for (o in " + poolField(r, s) + ") {")
		out.Indent()
		for _, f := range s.Fields {
			deserialize("o."+fieldName(f), 0, r, f.Type, out)
		}
		out.Dedent()
		out.WriteLine("}")
	}
	out.WriteLine("return true;")
	out.Dedent()
	out.WriteLine("}")

	out.Dedent()
	out.WriteLine("}")
}

func GenerateSources(input_file string, regions []*runtime.RegionSchema, output_dir string, pkg string, buffered fs.BufferedFileSystem) error {
	for _, r := range regions {
		for _, s := range r.Structs {
			outf := buffered.OutputFile(filepath.Join(output_dir, structName(s)+".hx"), 0644)
			ow, err := outf.GetWriter()
			if err != nil {
				return err
			}
			defer ow.Close()
			out := writer.MakeTabbedWriter("\t", ow)
			generateStruct(pkg, r, s, out)
		}
		outf := buffered.OutputFile(filepath.Join(output_dir, regionName(r)+".hx"), 0644)
		ow, err := outf.GetWriter()
		if err != nil {
			return err
		}
		defer ow.Close()
		out := writer.MakeTabbedWriter("\t", ow)
		generateRegion(pkg, r, out)
	}
	return nil
}
