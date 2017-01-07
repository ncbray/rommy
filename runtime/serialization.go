package runtime

import (
	"errors"
	"math"
)

func endOfData() error {
	return errors.New("end of data")
}

func outOfRange() error {
	return errors.New("value out of range")
}

type Serializer struct {
	data []byte
}

func MakeSerializer() *Serializer {
	return &Serializer{data: make([]byte, 0, 8)}
}

func (s *Serializer) Data() []byte {
	return s.data
}

func (s *Serializer) WriteBool(value bool) {
	var i uint8 = 0
	if value {
		i = 1
	}
	s.data = append(s.data, i)
}

func (s *Serializer) WriteUint8(value uint8) {
	s.data = append(s.data, value)
}

func (s *Serializer) WriteInt8(value int8) {
	s.WriteUint8(uint8(value))
}

func (s *Serializer) WriteUint16(value uint16) {
	s.data = append(s.data,
		uint8(value&0xff), uint8((value>>8)&0xff))
}

func (s *Serializer) WriteInt16(value int16) {
	s.WriteUint16(uint16(value))
}

func (s *Serializer) WriteUint32(value uint32) {
	s.data = append(s.data,
		uint8(value&0xff), uint8((value>>8)&0xff),
		uint8((value>>16)&0xff), uint8((value>>24)&0xff))
}

func (s *Serializer) WriteInt32(value int32) {
	s.WriteUint32(uint32(value))
}

func (s *Serializer) WriteUint64(value uint64) {
	s.data = append(s.data,
		uint8(value&0xff), uint8((value>>8)&0xff),
		uint8((value>>16)&0xff), uint8((value>>24)&0xff),
		uint8((value>>32)&0xff), uint8((value>>40)&0xff),
		uint8((value>>48)&0xff), uint8((value>>56)&0xff))
}

func (s *Serializer) WriteInt64(value int64) {
	s.WriteUint64(uint64(value))
}

func (s *Serializer) WriteIndex(index int, index_range int) error {
	if index < 0 || index >= index_range {
		return outOfRange()
	}
	if index_range <= 1 {
		// Implicit
	} else if index_range <= 1<<8 {
		s.WriteUint8(uint8(index))
	} else if index_range <= 1<<16 {
		s.WriteUint16(uint16(index))
	} else {
		s.WriteUint32(uint32(index))
	}
	return nil
}

func (s *Serializer) WriteCount(index int) error {
	if index < 0 || index > math.MaxInt32 {
		return outOfRange()
	}
	s.WriteUint32(uint32(index))
	return nil
}

func (s *Serializer) WriteString(value string) {
	s.WriteUint32(uint32(len(value)))
	s.data = append(s.data, value...)
}

type Deserializer struct {
	data []byte
}

func MakeDeserializer(data []byte) *Deserializer {
	return &Deserializer{data: data}
}

func (s *Deserializer) ReadBool() (bool, error) {
	v, err := s.ReadUint8()
	if err != nil {
		return false, err
	}
	if v > 1 {
		return false, outOfRange()
	}
	return v != 0, nil
}

func (s *Deserializer) ReadUint8() (uint8, error) {
	if len(s.data) >= 1 {
		b := s.data[0]
		s.data = s.data[1:]
		return b, nil
	} else {
		return 0, endOfData()
	}
}

func (s *Deserializer) ReadInt8() (int8, error) {
	v, err := s.ReadUint8()
	return int8(v), err
}

func (s *Deserializer) ReadUint16() (uint16, error) {
	if len(s.data) >= 2 {
		b := uint16(s.data[0]) | (uint16(s.data[1]) << 8)
		s.data = s.data[2:]
		return b, nil
	} else {
		return 0, endOfData()
	}
}

func (s *Deserializer) ReadInt16() (int16, error) {
	v, err := s.ReadUint16()
	return int16(v), err
}

func (s *Deserializer) ReadUint32() (uint32, error) {
	if len(s.data) >= 4 {
		b := uint32(s.data[0]) | (uint32(s.data[1]) << 8) | (uint32(s.data[2]) << 16) | (uint32(s.data[3]) << 24)
		s.data = s.data[4:]
		return b, nil
	} else {
		return 0, endOfData()
	}
}

func (s *Deserializer) ReadInt32() (int32, error) {
	v, err := s.ReadUint32()
	return int32(v), err
}

func (s *Deserializer) ReadUint64() (uint64, error) {
	if len(s.data) >= 8 {
		b := uint64(s.data[0]) | (uint64(s.data[1]) << 8) | (uint64(s.data[2]) << 16) | (uint64(s.data[3]) << 24) | (uint64(s.data[4]) << 32) | (uint64(s.data[5]) << 40) | (uint64(s.data[6]) << 48) | (uint64(s.data[7]) << 56)
		s.data = s.data[8:]
		return b, nil
	} else {
		return 0, endOfData()
	}
}

func (s *Deserializer) ReadInt64() (int64, error) {
	v, err := s.ReadUint64()
	return int64(v), err
}

func (s *Deserializer) ReadIndex(index_range int) (int, error) {
	var v int
	var err error
	if index_range <= 1 {
		v = 0
	} else if index_range <= 1<<8 {
		var p uint8
		p, err = s.ReadUint8()
		v = int(p)
	} else if index_range <= 1<<16 {
		var p uint16
		p, err = s.ReadUint16()
		v = int(p)
	} else {
		var p uint32
		p, err = s.ReadUint32()
		v = int(p)
	}
	if err != nil {
		return 0, err
	}
	if v >= index_range {
		return 0, outOfRange()
	}
	return v, err
}

func (s *Deserializer) ReadCount() (int, error) {
	p, err := s.ReadUint32()
	if err != nil {
		return 0, err
	}
	if p > math.MaxInt32 {
		return 0, outOfRange()
	}
	return int(p), err
}

func (s *Deserializer) ReadString() (string, error) {
	l, err := s.ReadUint32()
	if err != nil {
		return "", err
	}
	sl := int(l)
	if len(s.data) >= sl {
		v := s.data[0:sl]
		s.data = s.data[sl:]
		return string(v), nil
	} else {
		return "", endOfData()
	}
}
