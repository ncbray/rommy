package human

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

const (
	EndOfStream = '\uFFFD'
)

type RuneStreamPos int

type Location struct {
	file  string
	begin RuneStreamPos
	end   RuneStreamPos
}

type SourceString struct {
	Loc  Location
	Text string
}

type SourceInfo struct {
	File     string
	data     []byte
	lineInfo []RuneStreamPos
}

func (s *SourceInfo) Location(begin RuneStreamPos, end RuneStreamPos) Location {
	return Location{file: s.File, begin: begin, end: end}
}

func (s *SourceInfo) GetLineInfo(loc Location) (string, int, int, string) {
	// Lazy create
	if s.lineInfo == nil {
		info := []RuneStreamPos{}
		i := 0
		beginLine := true
		for i < len(s.data) {
			if beginLine {
				info = append(info, RuneStreamPos(i))
			}
			r, size := utf8.DecodeRune(s.data[i:])
			if r == utf8.RuneError {
				break
			}
			beginLine = r == '\n'
			i += size
		}
		info = append(info, RuneStreamPos(len(s.data)))
		s.lineInfo = info
	}
	// HACK linear scan
	for line, offset := range s.lineInfo {
		if loc.begin < offset {
			start := s.lineInfo[line-1]
			col_index := 0
			for i := start; i < loc.begin; {
				_, size := utf8.DecodeRune(s.data[i:])
				i += RuneStreamPos(size)
				col_index += 1
			}
			bytes := s.data[start:offset]
			return loc.file, line, col_index, string(bytes)
		}
	}

	return "", 0, 0, ""
}

type SourceSet struct {
	files map[string]*SourceInfo
}

func (s *SourceSet) Add(file string, data []byte) *SourceInfo {
	info := &SourceInfo{File: file, data: data}
	s.files[file] = info
	return info
}

func CreateSourceSet() *SourceSet {
	return &SourceSet{files: map[string]*SourceInfo{}}
}

type Status struct {
	Sources *SourceSet
	errors  int
}

func (s *Status) Error(loc Location, message string) {
	info := s.Sources.files[loc.file]
	file, line, col, text := info.GetLineInfo(loc)
	arrow := ""
	for i := 0; i < col; i++ {
		arrow += " "
	}
	arrow += "^"
	fmt.Printf("%s:%d:%d - ERROR %s\n%s%s\n", file, line, col, message, text, arrow)
	s.errors += 1
}

func (s *Status) ShouldStop() bool {
	return s.errors > 0
}

type RuneStream struct {
	info       *SourceInfo
	input      []byte
	currentPos RuneStreamPos
	nextPos    RuneStreamPos
	value      rune
}

func CreateRuneStream(info *SourceInfo, input []byte) *RuneStream {
	s := &RuneStream{info: info, input: input}
	s.Seek(0)
	return s
}

func (state *RuneStream) Position() RuneStreamPos {
	return state.currentPos
}

func (state *RuneStream) Seek(pos RuneStreamPos) {
	state.currentPos = pos
	r, size := utf8.DecodeRune(state.input[state.currentPos:])
	if r != utf8.RuneError {
		state.value = r
		state.nextPos = pos + RuneStreamPos(size)
	} else {
		state.value = EndOfStream
		state.nextPos = state.currentPos
	}
}

func (state *RuneStream) GetNext() {
	state.Seek(state.nextPos)
}

func (state *RuneStream) Peek() rune {
	return state.value
}

func (state *RuneStream) Is(value rune) bool {
	return state.value == value
}

func (state *RuneStream) IsLetter() bool {
	return unicode.IsLetter(state.value)
}

func (state *RuneStream) IsDigit() bool {
	return unicode.IsDigit(state.value)
}

func (state *RuneStream) IsSpace() bool {
	return unicode.IsSpace(state.value)
}

func (state *RuneStream) IsEndOfStream() bool {
	return state.value == EndOfStream
}

func (state *RuneStream) Slice(begin RuneStreamPos) SourceString {
	return SourceString{
		Loc:  state.info.Location(begin, state.currentPos),
		Text: string(state.input[begin:state.currentPos]),
	}
}

type RuneParserState struct {
	RuneStream
	deepest RuneStreamPos
	ok      bool
}

func (p *RuneParserState) Recover(pos RuneStreamPos) {
	if p.currentPos > p.deepest {
		p.deepest = p.currentPos
	}
	p.Seek(pos)
}

func (p *RuneParserState) Deepest() Location {
	if p.currentPos > p.deepest {
		p.deepest = p.currentPos
	}
	return p.info.Location(p.deepest, p.deepest+1)
}

func CreateRuneParser(info *SourceInfo, input []byte) *RuneParserState {
	s := &RuneParserState{
		*CreateRuneStream(info, input),
		0,
		true,
	}
	s.Seek(0)
	return s
}

type SimpleParser func(state *RuneParserState) bool

func Optional(state *RuneParserState, p SimpleParser) bool {
	pos := state.Position()
	ok := p(state)
	if !ok {
		state.Recover(pos)
	}
	return true
}

func Repeat(state *RuneParserState, p SimpleParser) bool {
	for {
		pos := state.Position()
		ok := p(state)
		if !ok {
			state.Recover(pos)
			return true
		}
	}
}
