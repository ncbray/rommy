package parser

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
