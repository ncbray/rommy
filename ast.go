package rommy

type Expr interface {
	isExpr()
}

type Integer struct {
	Raw SourceString
}

func (node *Integer) isExpr() {
}

type String struct {
	Raw   SourceString
	Value string
}

func (node *String) isExpr() {
}

type KeywordArg struct {
	Name  SourceString
	Value Expr
}

type TypeRef struct {
	Raw SourceString
}

type Struct struct {
	Type *TypeRef
	Loc  Location
	Args []*KeywordArg
}

func (node *Struct) isExpr() {
}

type List struct {
	Loc  Location
	Args []Expr
}

func (node *List) isExpr() {
}
