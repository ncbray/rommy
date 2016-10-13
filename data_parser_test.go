package rommy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCharacterDecoding(t *testing.T) {
	sources := CreateSourceSet()
	data := []byte("a1 ")
	info := sources.Add("test", data)
	state := CreateRuneStream(info, data)

	r := state.Peek()
	assert.Equal(t, 'a', r)
	assert.True(t, state.IsLetter())
	assert.False(t, state.IsDigit())
	assert.False(t, state.IsSpace())
	assert.False(t, state.IsEndOfStream())
	state.GetNext()

	r = state.Peek()
	assert.Equal(t, '1', r)
	assert.False(t, state.IsLetter())
	assert.True(t, state.IsDigit())
	assert.False(t, state.IsSpace())
	assert.False(t, state.IsEndOfStream())
	state.GetNext()

	r = state.Peek()
	assert.Equal(t, ' ', r)
	assert.False(t, state.IsLetter())
	assert.False(t, state.IsDigit())
	assert.True(t, state.IsSpace())
	assert.False(t, state.IsEndOfStream())
	state.GetNext()

	r = state.Peek()
	assert.Equal(t, EndOfStream, r)
	assert.False(t, state.IsLetter())
	assert.False(t, state.IsDigit())
	assert.False(t, state.IsSpace())
	assert.True(t, state.IsEndOfStream())
}

func TestParseInteger(t *testing.T) {
	sources := CreateSourceSet()
	status := &Status{Sources: sources}
	data := []byte("123")
	info := sources.Add("t", data)
	e := ParseData(info, data, status)
	assert.Equal(t, &Integer{
		Raw: SourceString{Loc: Location{file: "t", begin: 0, end: 3}, Text: "123"},
	}, e)
}

func TestParseConstructor(t *testing.T) {
	sources := CreateSourceSet()
	status := &Status{Sources: sources}
	data := []byte("A{\n  foo: [1, 2],\n  bar: B{baz: \"wot\\n\\\"m8t?\\\"\"}\n}")
	info := sources.Add("t", data)
	e := ParseData(info, data, status)
	assert.Equal(t, &Struct{
		Type: &TypeRef{Raw: SourceString{Loc: Location{file: "t", begin: 0, end: 1}, Text: "A"}},
		Loc:  Location{file: "t", begin: 1, end: 2},
		Args: []*KeywordArg{
			{
				Name: SourceString{Loc: Location{file: "t", begin: 5, end: 8}, Text: "foo"},
				Value: &List{
					Loc: Location{file: "t", begin: 10, end: 11},
					Args: []Expr{
						&Integer{Raw: SourceString{Loc: Location{file: "t", begin: 11, end: 12}, Text: "1"}},
						&Integer{Raw: SourceString{Loc: Location{file: "t", begin: 14, end: 15}, Text: "2"}},
					},
				},
			},
			{
				Name: SourceString{Loc: Location{file: "t", begin: 20, end: 23}, Text: "bar"},
				Value: &Struct{
					Type: &TypeRef{Raw: SourceString{Loc: Location{file: "t", begin: 25, end: 26}, Text: "B"}},
					Loc:  Location{file: "t", begin: 26, end: 27},
					Args: []*KeywordArg{
						{
							Name: SourceString{Loc: Location{file: "t", begin: 27, end: 30}, Text: "baz"},
							Value: &String{
								Raw:   SourceString{Loc: Location{file: "t", begin: 32, end: 47}, Text: "\"wot\\n\\\"m8t?\\\"\""},
								Value: "wot\n\"m8t?\"",
							},
						},
					},
				},
			},
		},
	}, e)
}
