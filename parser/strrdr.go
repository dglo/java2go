package parser

import (
	"io"
)

type StringReader struct {
	bytes []byte
	pos int
}

func NewStringReader(str string) *StringReader {
	return &StringReader{bytes: []byte(str), pos: 0}
}

func (rdr *StringReader) ReadByte() (byte, error) {
	if rdr.pos >= len(rdr.bytes) {
		return 0, io.EOF
	}

	c := rdr.bytes[rdr.pos]
	rdr.pos += 1
	return c, nil
}
