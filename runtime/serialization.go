package runtime

import (
	"encoding/binary"
	"errors"
	"io"
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
	s.WriteUint32(uint32(index))
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
	p, err := s.ReadUint32()
	v := int(p)
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

func ReadVarUint8(r io.ByteReader) (uint8, error) {
	value, err := binary.ReadUvarint(r)
	// TODO check range
	return uint8(value), err
}

func ReadVarUint16(r io.ByteReader) (uint16, error) {
	value, err := binary.ReadUvarint(r)
	// TODO check range
	return uint16(value), err
}

func ReadVarUint32(r io.ByteReader) (uint32, error) {
	value, err := binary.ReadUvarint(r)
	// TODO check range
	return uint32(value), err
}

func ReadVarUint64(r io.ByteReader) (uint64, error) {
	value, err := binary.ReadUvarint(r)
	return value, err
}

func ReadVarInt8(r io.ByteReader) (int8, error) {
	value, err := binary.ReadVarint(r)
	// TODO check range
	return int8(value), err
}

func ReadVarInt16(r io.ByteReader) (int16, error) {
	value, err := binary.ReadVarint(r)
	// TODO check range
	return int16(value), err
}

func ReadVarInt32(r io.ByteReader) (int32, error) {
	value, err := binary.ReadVarint(r)
	// TODO check range
	return int32(value), err
}

func ReadVarInt64(r io.ByteReader) (int64, error) {
	value, err := binary.ReadVarint(r)
	return value, err
}

func WriteVarUint8(value uint8, w io.Writer) error {
	return WriteVarUint64(uint64(value), w)
}

func WriteVarUint16(value uint16, w io.Writer) error {
	return WriteVarUint64(uint64(value), w)
}

func WriteVarUint32(value uint32, w io.Writer) error {
	return WriteVarUint64(uint64(value), w)
}

func WriteVarUint64(value uint64, w io.Writer) error {
	buf := make([]byte, 9)
	n := binary.PutUvarint(buf, value)
	_, err := w.Write(buf[:n])
	return err
}

func WriteVarInt8(value int8, w io.Writer) error {
	return WriteVarInt64(int64(value), w)
}

func WriteVarInt16(value int16, w io.Writer) error {
	return WriteVarInt64(int64(value), w)
}

func WriteVarInt32(value int32, w io.Writer) error {
	return WriteVarInt64(int64(value), w)
}

func WriteVarInt64(value int64, w io.Writer) error {
	buf := make([]byte, 9)
	n := binary.PutVarint(buf, value)
	_, err := w.Write(buf[:n])
	return err
}

func ReadString(r io.ByteReader) (string, error) {
	_, err := ReadVarUint32(r)
	if err != nil {
		return "", err
	}
	//buf, err := r.Read
	return "fake", nil
}

func WriteString(value string, w io.Writer) error {
	b := []byte(value)
	// TODO check for 4 GB strings
	err := WriteVarUint32(uint32(len(b)), w)
	if err != nil {
		return nil
	}
	_, err = w.Write(b)
	return err
}
