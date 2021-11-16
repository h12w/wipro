package wipro

type M interface {
	Marshal(*Writer)
	Unmarshal(Reader)
}

type Reader interface {
	ReadUint8() uint8
	ReadInt8() int8
	ReadInt16() int16
	ReadInt32() int32
	ReadUint32() uint32
	ReadUint64() uint64
	ReadInt64() int64
	ReadString() string
	ReadBytes() []byte

	Offset() int
	Err() error
	SetErr(error)
}
