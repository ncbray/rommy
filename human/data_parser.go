package human

import (
	"github.com/ncbray/rommy/parser"
)

func punc(state *parser.RuneParserState, value rune) bool {
	if state.Is(value) {
		state.GetNext()
		return true
	} else {
		return false
	}
}

func s(state *parser.RuneParserState) {
	for state.IsSpace() {
		state.GetNext()
	}
}

func identifier(state *parser.RuneParserState) (parser.SourceString, bool) {
	if state.IsLetter() || state.Is('_') {
		begin := state.Position()
		state.GetNext()
		for state.IsLetter() || state.IsDigit() || state.Is('_') {
			state.GetNext()
		}
		return state.Slice(begin), true
	} else {
		return parser.SourceString{}, false
	}
}

func parseKeywordArg(state *parser.RuneParserState) (*KeywordArg, bool) {
	name, ok := identifier(state)
	if !ok {
		return nil, false
	}
	s(state)
	if !state.Is(':') {
		return nil, false
	}
	state.GetNext()
	s(state)

	expr, ok := parseExpr(state)
	if !ok {
		return nil, false
	}
	return &KeywordArg{Name: name, Value: expr}, true
}

func parseBoolean(state *parser.RuneParserState) (*Boolean, bool) {
	p := state.Position()
	if state.Is('t') {
		state.GetNext()
		if !state.Is('r') {
			return nil, false
		}
		state.GetNext()
		if !state.Is('u') {
			return nil, false
		}
		state.GetNext()
		if !state.Is('e') {
			return nil, false
		}
		state.GetNext()
		return &Boolean{Loc: state.Slice(p).Loc, Value: true}, true
	} else if state.Is('f') {
		state.GetNext()
		if !state.Is('a') {
			return nil, false
		}
		state.GetNext()
		if !state.Is('l') {
			return nil, false
		}
		state.GetNext()
		if !state.Is('s') {
			return nil, false
		}
		state.GetNext()
		if !state.Is('e') {
			return nil, false
		}
		state.GetNext()
		return &Boolean{Loc: state.Slice(p).Loc, Value: false}, true
	}
	return nil, false
}

func parseString(state *parser.RuneParserState) (*String, bool) {
	begin := state.Position()
	if !punc(state, '"') {
		return nil, false
	}
	value := []rune{}
	for !state.IsEndOfStream() && !state.Is('"') {
		current := state.Peek()
		state.GetNext()
		if current == '\\' {
			current = state.Peek()
			state.GetNext()
			switch current {
			case '"', '\\':
				// Pass through
			case 'n':
				current = '\n'
			case 't':
				current = '\t'
			default:
				return nil, false
			}
		}
		value = append(value, current)
	}
	if !punc(state, '"') {
		return nil, false
	}
	return &String{Raw: state.Slice(begin), Value: string(value)}, true
}

func parseKeywordArgList(state *parser.RuneParserState) ([]*KeywordArg, bool) {
	args := []*KeywordArg{}
	parser.Optional(state, func(state *parser.RuneParserState) bool {
		arg, ok := parseKeywordArg(state)
		if !ok {
			return false
		}
		args = append(args, arg)
		parser.Repeat(state, func(state *parser.RuneParserState) bool {
			s(state)
			if !punc(state, ',') {
				return false
			}
			s(state)
			arg, ok := parseKeywordArg(state)
			if !ok {
				return false
			}
			args = append(args, arg)
			return true
		})
		// Trailing comma
		parser.Optional(state, func(state *parser.RuneParserState) bool {
			s(state)
			if !punc(state, ',') {
				return false
			}
			return true
		})
		return true
	})
	return args, true
}

func optionalTypeRef(state *parser.RuneParserState) (*TypeRef, bool) {
	var name parser.SourceString
	var ok bool
	parser.Optional(state, func(state *parser.RuneParserState) bool {
		name, ok = identifier(state)
		return ok
	})
	if ok {
		return &TypeRef{Raw: name}, true
	} else {
		return nil, false
	}
}

func parseStruct(state *parser.RuneParserState) (*Struct, bool) {
	t, ok := optionalTypeRef(state)
	if ok {
		s(state)
	}
	begin := state.Position()
	if !punc(state, '{') {
		return nil, false
	}
	loc := state.Slice(begin).Loc
	s(state)
	args, ok := parseKeywordArgList(state)
	if !ok {
		return nil, false
	}
	s(state)
	if !punc(state, '}') {
		return nil, false
	}
	return &Struct{Type: t, Loc: loc, Args: args}, true
}

func parseList(state *parser.RuneParserState) (*List, bool) {
	begin := state.Position()
	if !punc(state, '[') {
		return nil, false
	}
	loc := state.Slice(begin).Loc
	s(state)

	args := []Expr{}
	parser.Optional(state, func(state *parser.RuneParserState) bool {
		arg, ok := parseExpr(state)
		if !ok {
			return false
		}
		args = append(args, arg)
		parser.Repeat(state, func(state *parser.RuneParserState) bool {
			s(state)
			if !punc(state, ',') {
				return false
			}
			s(state)
			arg, ok := parseExpr(state)
			if !ok {
				return false
			}
			args = append(args, arg)
			return true
		})

		// Trailing comma
		parser.Optional(state, func(state *parser.RuneParserState) bool {
			s(state)
			if !punc(state, ',') {
				return false
			}
			return true
		})
		return true
	})
	s(state)
	if !punc(state, ']') {
		return nil, false
	}
	return &List{Loc: loc, Args: args}, true
}

func parseExpr(state *parser.RuneParserState) (Expr, bool) {
	begin := state.Position()
	switch {
	case state.IsDigit():
		state.GetNext()
		for state.IsDigit() {
			state.GetNext()
		}
		return &Integer{Raw: state.Slice(begin)}, true
	case state.Is('t') || state.Is('f'):
		// HACK types that start with "t" or "f" will choke.
		return parseBoolean(state)
	case state.IsLetter() || state.Is('_'):
		return parseStruct(state)
	case state.Is('{'):
		return parseStruct(state)
	case state.Is('"'):
		return parseString(state)
	case state.Is('['):
		return parseList(state)
	default:
		return nil, false
	}
}

// Parse text into an AST.
func ParseData(info *parser.SourceInfo, input []byte, status *parser.Status) Expr {
	state := parser.CreateRuneParser(info, input)
	s(state)
	e, ok := parseExpr(state)
	if ok {
		s(state)
	}
	if !ok || !state.IsEndOfStream() {
		status.Error(state.Deepest(), "unexpected character")
	}
	return e
}
