package rommy

func Punc(state *RuneParserState, value rune) bool {
	if state.Is(value) {
		state.GetNext()
		return true
	} else {
		return false
	}
}

func S(state *RuneParserState) {
	for state.IsSpace() {
		state.GetNext()
	}
}

func Identifier(state *RuneParserState) (SourceString, bool) {
	if state.IsLetter() || state.Is('_') {
		begin := state.Position()
		state.GetNext()
		for state.IsLetter() || state.IsDigit() || state.Is('_') {
			state.GetNext()
		}
		return state.Slice(begin), true
	} else {
		return SourceString{}, false
	}
}

func ParseKeywordArg(state *RuneParserState) (*KeywordArg, bool) {
	name, ok := Identifier(state)
	if !ok {
		return nil, false
	}
	S(state)
	if !state.Is(':') {
		return nil, false
	}
	state.GetNext()
	S(state)

	expr, ok := ParseExpr(state)
	if !ok {
		return nil, false
	}
	return &KeywordArg{Name: name, Value: expr}, true
}

func ParseString(state *RuneParserState) (*String, bool) {
	begin := state.Position()
	if !Punc(state, '"') {
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
	if !Punc(state, '"') {
		return nil, false
	}
	return &String{Raw: state.Slice(begin), Value: string(value)}, true
}

func ParseKeywordArgList(state *RuneParserState) ([]*KeywordArg, bool) {
	args := []*KeywordArg{}
	Optional(state, func(state *RuneParserState) bool {
		arg, ok := ParseKeywordArg(state)
		if !ok {
			return false
		}
		args = append(args, arg)
		Repeat(state, func(state *RuneParserState) bool {
			S(state)
			if !Punc(state, ',') {
				return false
			}
			S(state)
			arg, ok := ParseKeywordArg(state)
			if !ok {
				return false
			}
			args = append(args, arg)
			return true
		})
		// Trailing comma
		Optional(state, func(state *RuneParserState) bool {
			S(state)
			if !Punc(state, ',') {
				return false
			}
			return true
		})
		return true
	})
	return args, true
}

func OptionalTypeRef(state *RuneParserState) (*TypeRef, bool) {
	var name SourceString
	var ok bool
	Optional(state, func(state *RuneParserState) bool {
		name, ok = Identifier(state)
		return ok
	})
	if ok {
		return &TypeRef{Raw: name}, true
	} else {
		return nil, false
	}
}

func ParseStruct(state *RuneParserState) (*Struct, bool) {
	t, ok := OptionalTypeRef(state)
	if ok {
		S(state)
	}
	begin := state.Position()
	if !Punc(state, '{') {
		return nil, false
	}
	loc := state.Slice(begin).Loc
	S(state)
	args, ok := ParseKeywordArgList(state)
	if !ok {
		return nil, false
	}
	S(state)
	if !Punc(state, '}') {
		return nil, false
	}
	return &Struct{Type: t, Loc: loc, Args: args}, true
}

func ParseList(state *RuneParserState) (*List, bool) {
	begin := state.Position()
	if !Punc(state, '[') {
		return nil, false
	}
	loc := state.Slice(begin).Loc
	S(state)

	args := []Expr{}
	Optional(state, func(state *RuneParserState) bool {
		arg, ok := ParseExpr(state)
		if !ok {
			return false
		}
		args = append(args, arg)
		Repeat(state, func(state *RuneParserState) bool {
			S(state)
			if !Punc(state, ',') {
				return false
			}
			S(state)
			arg, ok := ParseExpr(state)
			if !ok {
				return false
			}
			args = append(args, arg)
			return true
		})

		// Trailing comma
		Optional(state, func(state *RuneParserState) bool {
			S(state)
			if !Punc(state, ',') {
				return false
			}
			return true
		})
		return true
	})
	S(state)
	if !Punc(state, ']') {
		return nil, false
	}
	return &List{Loc: loc, Args: args}, true
}

func ParseExpr(state *RuneParserState) (Expr, bool) {
	begin := state.Position()
	switch {
	case state.IsDigit():
		state.GetNext()
		for state.IsDigit() {
			state.GetNext()
		}
		return &Integer{Raw: state.Slice(begin)}, true
	case state.IsLetter() || state.Is('_'):
		return ParseStruct(state)
	case state.Is('{'):
		return ParseStruct(state)
	case state.Is('"'):
		return ParseString(state)
	case state.Is('['):
		return ParseList(state)
	default:
		return nil, false
	}
}

func ParseData(info *SourceInfo, input []byte, s *Status) Expr {
	state := CreateRuneParser(info, input)
	S(state)
	e, ok := ParseExpr(state)
	if ok {
		S(state)
	}
	if !ok || !state.IsEndOfStream() {
		s.Error(state.Deepest(), "unexpected character")
	}
	return e
}
