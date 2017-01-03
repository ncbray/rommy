package human

import (
	"github.com/ncbray/compilerutil/writer"
	"io"
)

func isSimple(expr Expr) bool {
	switch expr := expr.(type) {
	case *String, *Integer:
		return true
	case *Struct:
		if len(expr.Args) >= 6 {
			return false
		}
		for _, arg := range expr.Args {
			if !isSimple(arg.Value) {
				return false
			}
		}
		return true
	case *List:
		if len(expr.Args) > 1 {
			return false
		}
		for _, arg := range expr.Args {
			if !isSimple(arg) {
				return false
			}
		}
		return true
	default:
		panic(expr)
	}
}

func isDefault(expr Expr) bool {
	switch expr := expr.(type) {
	case *String:
		return expr.Value == ""
	case *Integer:
		// HACK
		return expr.Raw.Text == "0"
	case *Struct:
		return false
	default:
		panic(expr)
	}
}

func writeExpr(expr Expr, out *writer.TabbedWriter) {
	switch expr := expr.(type) {
	case *String:
		out.WriteString(expr.Raw.Text)
	case *Integer:
		out.WriteString(expr.Raw.Text)
	case *List:
		one_line := isSimple(expr)

		//if expr.Type != "" {
		//	out.WriteString(expr.Type)
		//	out.WriteString(" ")
		//}
		out.WriteString("[")
		if !one_line {
			out.EndOfLine()
			out.Indent()
		}
		for i, arg := range expr.Args {
			writeExpr(arg, out)
			if one_line {
				if i < len(expr.Args)-1 {
					out.WriteString(", ")
				}
			} else {
				out.WriteString(",")
				out.EndOfLine()
			}
		}
		if !one_line {
			out.Dedent()
		}
		out.WriteString("]")
	case *Struct:
		one_line := isSimple(expr)
		if expr.Type != nil {
			out.WriteString(expr.Type.Raw.Text)
			out.WriteString(" ")
		}
		out.WriteString("{")

		if one_line {
			for i, arg := range expr.Args {
				out.WriteString(arg.Name.Text)
				out.WriteString(": ")
				writeExpr(arg.Value, out)
				if i < len(expr.Args)-1 {
					out.WriteString(", ")
				}
			}
		} else {
			out.EndOfLine()
			out.Indent()
			for _, arg := range expr.Args {
				if isDefault(arg.Value) {
					continue
				}
				out.WriteString(arg.Name.Text)
				out.WriteString(": ")
				writeExpr(arg.Value, out)
				out.WriteString(",")
				out.EndOfLine()
			}
			out.Dedent()
		}
		out.WriteString("}")
	default:
		panic(expr)
	}
}

func Write(expr Expr, w io.Writer) {
	out := writer.MakeTabbedWriter("  ", w)
	writeExpr(expr, out)
	out.EndOfLine()
}
