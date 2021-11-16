package wipro

import (
	"errors"
	"io"
)

var (
	ErrPrefix        = "proto: "
	ErrUnexpectedEOF = errors.New(ErrPrefix + "unexpected EOF")
)

func Send(m M, conn io.Writer) error {
	var w Writer
	m.Marshal(&w)
	if _, err := conn.Write(w.B); err != nil {
		return err
	}
	return nil
}

func Receive(conn io.Reader, m M) error {
	r := WireReader{B: make([]byte, 4)}
	if _, err := conn.Read(r.B); err != nil {
		return err
	}
	size := int(r.ReadInt32())
	r.Grow(size)
	if size > 0 {
		if _, err := io.ReadAtLeast(conn, r.B[4:], size); err != nil {
			return err
		}
	}
	r.Reset()
	m.Unmarshal(&r)
	return r.err
}

type Writer struct {
	B []byte
}

type WireReader struct {
	B      []byte
	offset int
	err    error
}

func (r *WireReader) Offset() int {
	return r.offset
}

func (r *WireReader) Err() error {
	return r.err
}

func (r *WireReader) SetErr(err error) {
	if err == nil {
		// NOTE: server optimization, ignore err for sized array element and quit silently
		r.offset = len(r.B)
	}
	r.err = err
}

func (w *Writer) WriteUint8(i uint8) {
	w.B = append(w.B, i)
}

func (r *WireReader) ReadUint8() uint8 {
	if r.err != nil {
		return 0
	}
	i := r.offset
	if i+1 > len(r.B) {
		r.err = ErrUnexpectedEOF
		return 0
	}
	r.offset++
	return r.B[i]
}

func (w *Writer) WriteInt8(i int8) {
	w.WriteUint8(uint8(i))
}

func (r *WireReader) ReadInt8() int8 {
	return int8(r.ReadUint8())
}

func (w *Writer) WriteInt16(i int16) {
	w.B = append(w.B, byte(i>>8), byte(i))
}

func (r *WireReader) ReadInt16() int16 {
	if r.err != nil {
		return 0
	}
	i := r.offset
	if i+2 > len(r.B) {
		r.err = ErrUnexpectedEOF
		return 0
	}
	r.offset += 2
	return int16(r.B[i])<<8 | int16(r.B[i+1])
}

func (w *Writer) WriteInt32(i int32) {
	w.B = append(w.B, byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

func (w *Writer) WriteUint32(i uint32) {
	w.B = append(w.B, byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

func (r *WireReader) ReadInt32() int32 {
	if r.err != nil {
		return 0
	}
	i := r.offset
	if i+4 > len(r.B) {
		r.err = ErrUnexpectedEOF
		return 0
	}
	r.offset += 4
	return int32(r.B[i])<<24 | int32(r.B[i+1])<<16 | int32(r.B[i+2])<<8 | int32(r.B[i+3])
}

func (r *WireReader) ReadUint32() uint32 {
	if r.err != nil {
		return 0
	}
	i := r.offset
	if i+4 > len(r.B) {
		r.err = ErrUnexpectedEOF
		return 0
	}
	r.offset += 4
	return uint32(r.B[i])<<24 | uint32(r.B[i+1])<<16 | uint32(r.B[i+2])<<8 | uint32(r.B[i+3])
}

func (w *Writer) WriteUint64(i uint64) {
	w.B = append(w.B, byte(i>>56), byte(i>>48), byte(i>>40), byte(i>>32), byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

func (r *WireReader) ReadUint64() uint64 {
	if r.err != nil {
		return 0
	}
	i := r.offset
	if i+8 > len(r.B) {
		r.err = ErrUnexpectedEOF
		return 0
	}
	r.offset += 8
	return uint64(r.B[i])<<56 | uint64(r.B[i+1])<<48 | uint64(r.B[i+2])<<40 | uint64(r.B[i+3])<<32 |
		uint64(r.B[i+4])<<24 | uint64(r.B[i+5])<<16 | uint64(r.B[i+6])<<8 | uint64(r.B[i+7])
}

func (w *Writer) WriteInt64(i int64) {
	w.WriteUint64(uint64(i))
}

func (r *WireReader) ReadInt64() int64 {
	return int64(r.ReadUint64())
}

func (w *Writer) WriteString(s string) {
	w.WriteInt16(int16(len(s)))
	w.B = append(w.B, s...)
}

func (r *WireReader) ReadString() string {
	if r.err != nil {
		return ""
	}
	l := int(r.ReadInt16())
	if l <= 0 {
		return ""
	}
	i := r.offset
	if i+l > len(r.B) {
		r.err = ErrUnexpectedEOF
		return ""
	}
	r.offset += l
	return string(r.B[i : i+l])
}

func (w *Writer) WriteBytes(bs []byte) {
	w.WriteInt32(int32(len(bs)))
	w.B = append(w.B, bs...)
}

func (r *WireReader) ReadBytes() []byte {
	if r.err != nil {
		return nil
	}
	l := int(r.ReadInt32())
	if l <= 0 {
		return nil
	}
	i := r.offset
	if i+l > len(r.B) {
		r.err = ErrUnexpectedEOF
		return nil
	}
	r.offset += l
	return r.B[i : i+l]
}

func (w *Writer) SetInt32(offset int, i int32) {
	w.B[offset] = byte(i >> 24)
	w.B[offset+1] = byte(i >> 16)
	w.B[offset+2] = byte(i >> 8)
	w.B[offset+3] = byte(i)
}

func (w *Writer) SetUint32(offset int, i uint32) {
	w.B[offset] = byte(i >> 24)
	w.B[offset+1] = byte(i >> 16)
	w.B[offset+2] = byte(i >> 8)
	w.B[offset+3] = byte(i)
}

func (r *WireReader) Grow(n int) {
	if n <= 0 {
		return
	}
	b := make([]byte, len(r.B)+n)
	copy(b, r.B)
	r.B = b
}

func (r *WireReader) Reset() {
	r.offset = 0
}
