package runtime

import (
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
)

func TestUint8(t *testing.T) {
	s := MakeSerializer()
	var i, min, max uint8
	min = 0
	max = math.MaxUint8
	i = min
	for {
		s.WriteUint8(i)
		if i == max {
			break
		}
		i++
	}
	data := s.Data()
	assert.Equal(t, (int(max)-int(min)+1)*1, len(data))
	d := MakeDeserializer(data)
	i = min
	for {
		v, err := d.ReadUint8()
		assert.Nil(t, err)
		assert.Equal(t, i, v)
		if t.Failed() {
			return
		}
		if i == max {
			break
		}
		i++
	}
	_, err := d.ReadUint8()
	assert.NotNil(t, err)
}

func TestInt8(t *testing.T) {
	s := MakeSerializer()
	var i, min, max int8
	min = math.MinInt8
	max = math.MaxInt8
	i = min
	for {
		s.WriteInt8(i)
		if i == max {
			break
		}
		i++
	}
	data := s.Data()
	assert.Equal(t, (int(max)-int(min)+1)*1, len(data))
	d := MakeDeserializer(data)
	i = math.MinInt8
	for {
		v, err := d.ReadInt8()
		assert.Nil(t, err)
		assert.Equal(t, i, v)
		if t.Failed() {
			return
		}
		if i == math.MaxInt8 {
			break
		}
		i++
	}
	_, err := d.ReadInt8()
	assert.NotNil(t, err)
}

func TestUint16(t *testing.T) {
	s := MakeSerializer()
	var i, min, max uint16
	min = 0
	max = math.MaxUint16
	i = min
	for {
		s.WriteUint16(i)
		if i == max {
			break
		}
		i++
	}
	data := s.Data()
	assert.Equal(t, (int(max)-int(min)+1)*2, len(data))
	d := MakeDeserializer(data)
	i = min
	for {
		v, err := d.ReadUint16()
		assert.Nil(t, err)
		assert.Equal(t, i, v)
		if t.Failed() {
			return
		}
		if i == max {
			break
		}
		i++
	}
	_, err := d.ReadUint16()
	assert.NotNil(t, err)
}

func TestInt16(t *testing.T) {
	s := MakeSerializer()
	var i, min, max int16
	min = math.MinInt16
	max = math.MaxInt16
	i = min
	for {
		s.WriteInt16(i)
		if i == max {
			break
		}
		i++
	}
	data := s.Data()
	assert.Equal(t, (int(max)-int(min)+1)*2, len(data))
	d := MakeDeserializer(data)
	i = math.MinInt16
	for {
		v, err := d.ReadInt16()
		assert.Nil(t, err)
		assert.Equal(t, i, v)
		if t.Failed() {
			return
		}
		if i == math.MaxInt16 {
			break
		}
		i++
	}
	_, err := d.ReadInt16()
	assert.NotNil(t, err)
}

func TestUint32(t *testing.T) {
	s := MakeSerializer()
	expected := []uint32{}
	for i := 0; i < 1024; i++ {
		expected = append(expected, rand.Uint32())
	}
	for _, value := range expected {
		s.WriteUint32(value)
	}
	data := s.Data()
	assert.Equal(t, len(expected)*4, len(data))
	d := MakeDeserializer(data)
	for _, value := range expected {
		actual, err := d.ReadUint32()
		assert.Nil(t, err)
		assert.Equal(t, value, actual)
		if t.Failed() {
			return
		}
	}
	_, err := d.ReadUint32()
	assert.NotNil(t, err)
}

func TestInt32(t *testing.T) {
	s := MakeSerializer()
	expected := []int32{}
	for i := 0; i < 1024; i++ {
		expected = append(expected, int32(rand.Uint32()))
	}
	for _, value := range expected {
		s.WriteInt32(value)
	}
	data := s.Data()
	assert.Equal(t, len(expected)*4, len(data))
	d := MakeDeserializer(data)
	for _, value := range expected {
		actual, err := d.ReadInt32()
		assert.Nil(t, err)
		assert.Equal(t, value, actual)
		if t.Failed() {
			return
		}
	}
	_, err := d.ReadInt32()
	assert.NotNil(t, err)
}

func TestUint64(t *testing.T) {
	s := MakeSerializer()
	expected := []uint64{}
	for i := 0; i < 1024; i++ {
		expected = append(expected, uint64(rand.Uint32())|(uint64(rand.Uint32())<<32))
	}
	for _, value := range expected {
		s.WriteUint64(value)
	}
	data := s.Data()
	assert.Equal(t, len(expected)*8, len(data))
	d := MakeDeserializer(data)
	for _, value := range expected {
		actual, err := d.ReadUint64()
		assert.Nil(t, err)
		assert.Equal(t, value, actual)
		if t.Failed() {
			return
		}
	}
	_, err := d.ReadUint64()
	assert.NotNil(t, err)
}

func TestInt64(t *testing.T) {
	s := MakeSerializer()
	expected := []int64{}
	for i := 0; i < 1024; i++ {
		expected = append(expected, int64(uint64(rand.Uint32())|(uint64(rand.Uint32())<<32)))
	}
	for _, value := range expected {
		s.WriteInt64(value)
	}
	data := s.Data()
	assert.Equal(t, len(expected)*8, len(data))
	d := MakeDeserializer(data)
	for _, value := range expected {
		actual, err := d.ReadInt64()
		assert.Nil(t, err)
		assert.Equal(t, value, actual)
		if t.Failed() {
			return
		}
	}
	_, err := d.ReadInt64()
	assert.NotNil(t, err)
}

func TestString(t *testing.T) {
	s := MakeSerializer()
	expected := []string{"foo", "bar", "baz"}
	for _, value := range expected {
		s.WriteString(value)
	}
	data := s.Data()
	d := MakeDeserializer(data)
	for _, value := range expected {
		actual, err := d.ReadString()
		assert.Nil(t, err)
		assert.Equal(t, value, actual)
		if t.Failed() {
			return
		}
	}
	_, err := d.ReadString()
	assert.NotNil(t, err)
}
