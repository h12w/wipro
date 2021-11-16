package wipro

import "io"

type FileReader struct {
	r io.Reader
}

func NewFileReader(r io.Reader) *FileReader {
	return &FileReader{r: r}
}

func (r *FileReader) ReadUint8() uint8 {
	panic("not implemented")
}
func (r *FileReader) ReadInt8() int8 {
	panic("not implemented")
}

func (r *FileReader) ReadInt16() int16 {
	panic("not implemented")
}

func (r *FileReader) ReadInt32() int32 {
	panic("not implemented")
}

func (r *FileReader) ReadUint32() uint32 {
	panic("not implemented")
}

func (r *FileReader) ReadUint64() uint64 {
	panic("not implemented")
}

func (r *FileReader) ReadInt64() int64 {
	panic("not implemented")
}
func (r *FileReader) ReadString() string {
	panic("not implemented")
}
func (r *FileReader) ReadBytes() []byte {
	panic("not implemented")
}
func (r *FileReader) Offset() int {
	panic("not implemented")
}
func (r *FileReader) Err() error {
	panic("not implemented")
}
func (r *FileReader) SetErr(error) {
	panic("not implemented")
}
