package human

import (
	"github.com/ncbray/rommy/parser"
)

type Expr interface {
	isExpr()
}

type Integer struct {
	Raw parser.SourceString
}

func (node *Integer) isExpr() {
}

type Boolean struct {
	Loc   parser.Location
	Value bool
}

func (node *Boolean) isExpr() {
}

type String struct {
	Raw   parser.SourceString
	Value string
}

func (node *String) isExpr() {
}

type KeywordArg struct {
	Name  parser.SourceString
	Value Expr
}

type TypeRef struct {
	Raw parser.SourceString
}

type Struct struct {
	Type *TypeRef
	Loc  parser.Location
	Args []*KeywordArg
}

func (node *Struct) isExpr() {
}

type List struct {
	Loc  parser.Location
	Args []Expr
}

func (node *List) isExpr() {
}
