package human

import (
	"github.com/ncbray/rommy/parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseInteger(t *testing.T) {
	sources := parser.CreateSourceSet()
	status := &parser.Status{Sources: sources}
	data := []byte("123")
	info := sources.Add("t", data)
	e := ParseData(info, data, status)
	assert.Equal(t, &Integer{
		Raw: parser.SourceString{Loc: info.Location(0, 3), Text: "123"},
	}, e)
}

func TestParseConstructor(t *testing.T) {
	sources := parser.CreateSourceSet()
	status := &parser.Status{Sources: sources}
	data := []byte("A{\n  foo: [1, 2],\n  bar: B{baz: \"wot\\n\\\"m8t?\\\"\"}\n}")
	info := sources.Add("t", data)
	e := ParseData(info, data, status)
	assert.Equal(t, &Struct{
		Type: &TypeRef{Raw: parser.SourceString{Loc: info.Location(0, 1), Text: "A"}},
		Loc:  info.Location(1, 2),
		Args: []*KeywordArg{
			{
				Name: parser.SourceString{Loc: info.Location(5, 8), Text: "foo"},
				Value: &List{
					Loc: info.Location(10, 11),
					Args: []Expr{
						&Integer{Raw: parser.SourceString{Loc: info.Location(11, 12), Text: "1"}},
						&Integer{Raw: parser.SourceString{Loc: info.Location(14, 15), Text: "2"}},
					},
				},
			},
			{
				Name: parser.SourceString{Loc: info.Location(20, 23), Text: "bar"},
				Value: &Struct{
					Type: &TypeRef{Raw: parser.SourceString{Loc: info.Location(25, 26), Text: "B"}},
					Loc:  info.Location(26, 27),
					Args: []*KeywordArg{
						{
							Name: parser.SourceString{Loc: info.Location(27, 30), Text: "baz"},
							Value: &String{
								Raw:   parser.SourceString{Loc: info.Location(32, 47), Text: "\"wot\\n\\\"m8t?\\\"\""},
								Value: "wot\n\"m8t?\"",
							},
						},
					},
				},
			},
		},
	}, e)
}
