package runtime

import (
	"encoding/binary"
	"io"
)

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
	buf := make([]byte, 16)
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
	buf := make([]byte, 16)
	n := binary.PutVarint(buf, value)
	_, err := w.Write(buf[:n])
	return err
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
