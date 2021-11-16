package gen

import (
	"fmt"
	"io"
	"strings"

	"h12.io/gengo"
)

func (t GoTypes) GoFuncs(w io.Writer, packageName string) {
	fpl(w, "package "+packageName)
	fpl(w, "import (")
	if t.hasCRC() {
		fpl(w, `"hash/crc32"`)
	}
	fpl(w, `"h12.io/wipro"`)
	fpl(w, ")")
	for _, decl := range t.TypeDecls {
		genMarshalFunc(w, decl)
		fpl(w, "")
		genUnmarshalFunc(w, decl)
		fpl(w, "")
	}
}

func (t GoTypes) hasCRC() bool {
	for _, decl := range t.TypeDecls {
		if hasCRC(decl) {
			return true
		}
	}
	return false
}

func hasCRC(decl *gengo.TypeDecl) bool {
	t := &decl.Type
	return t.Kind == gengo.StructKind &&
		len(t.Fields) > 0 && t.Fields[0].Name == "CRC"
}

func genMarshalFunc(w io.Writer, decl *gengo.TypeDecl) {
	t := &decl.Type
	if t.Kind == gengo.IdentKind && t.Ident == "wipro.M" {
		return
	}
	fpl(w, "func (t *%s) Marshal(w *wipro.Writer) {", decl.Name)
	switch t.Kind {
	case gengo.StructKind:
		if len(t.Fields) > 0 {
			if f0 := t.Fields[0]; f0.Name == "Size" || f0.Name == "CRC" {
				fpl(w, "offset := len(w.B)")
				marshalField(w, t.Fields[0])
				fpl(w, "start := len(w.B)")
				for _, field := range t.Fields[1:] {
					marshalField(w, field)
				}
				switch f0.Name {
				case "Size":
					fpl(w, "w.SetInt32(offset, int32(len(w.B)-start))")
				case "CRC":
					fpl(w, "w.SetUint32(offset, crc32.ChecksumIEEE(w.B[start:]))")
				}
			} else {
				for _, f := range t.Fields {
					marshalField(w, f)
				}
			}
		} else {
			fpl(w, "// no fields for type %s, %v", decl.Name, decl.Type)
		}
	case gengo.ArrayKind:
		marshalValue(w, "(*t)", t, t.Ident)
	case gengo.IdentKind:
		marshalValue(w, "(*t)", t, decl.Name)
	default:
		fpl(w, "// type %s, %v", decl.Name, decl.Type)
	}
	fpl(w, "}")
}

func genUnmarshalFunc(w io.Writer, decl *gengo.TypeDecl) {
	t := &decl.Type
	if t.Kind == gengo.IdentKind && t.Ident == "wipro.M" {
		return
	}
	fpl(w, "func (t *%s) Unmarshal(r wipro.Reader) {", decl.Name)
	switch t.Kind {
	case gengo.StructKind:
		if len(t.Fields) > 0 {
			if f0 := t.Fields[0]; f0.Name == "Size" || f0.Name == "CRC" {
				unmarshalField(w, t.Fields[0])
				fpl(w, "start := r.Offset()")
				for _, field := range t.Fields[1:] {
					unmarshalField(w, field)
				}
				switch f0.Name {
				case "Size":
					fpl(w, "if r.Err() == nil && int(t.Size) != r.Offset()-start {")
					fpl(w, `r.SetErr(ErrSizeMismatch)`)
					fpl(w, "}")
				case "CRC":
					fpl(w, "// if r.Err() == nil && t.CRC != crc32.ChecksumIEEE(r.B[start:r.Offset()]) {")
					fpl(w, `// r.SetErr(ErrCRCMismatch)`)
					fpl(w, "// }")
					fpl(w, "_ = start")
				}
			} else {
				for _, f := range t.Fields {
					unmarshalField(w, f)
				}
			}
		} else {
			fpl(w, "// no fields for type %s, %v", decl.Name, decl.Type)
		}
	case gengo.ArrayKind:
		unmarshalValue(w, "(*t)", t, t.Ident)
	case gengo.IdentKind:
		unmarshalValue(w, "(*t)", t, decl.Name)
	default:
		fpl(w, "// type %s, %v", decl.Name, decl.Type)
	}
	fpl(w, "}")
}

func marshalField(w io.Writer, f *gengo.Field) {
	fName := "t." + f.Name
	if f.Name == "" {
		fName = "t." + embeddedVarName(f.Type.Ident)
	}
	marshalValue(w, fName, &f.Type, f.Type.Ident)
}

func unmarshalField(w io.Writer, f *gengo.Field) {
	fName := "t." + f.Name
	if f.Name == "" {
		fName = "t." + embeddedVarName(f.Type.Ident)
	}
	unmarshalValue(w, fName, &f.Type, f.Type.Ident)
}

func embeddedVarName(s string) string {
	items := strings.Split(s, ".")
	return items[len(items)-1]
}

func marshalValue(w io.Writer, name string, typ *gengo.Type, declType string) {
	switch typ.Kind {
	case gengo.IdentKind:
		conv := ""
		if typ.Ident != declType {
			conv = declType
		}
		switch typ.Ident {
		case "int64":
			marshalInt(w, name, 64, conv)
		case "int32":
			marshalInt(w, name, 32, conv)
		case "int16":
			marshalInt(w, name, 16, conv)
		case "int8":
			marshalInt(w, name, 8, conv)
		case "uint64":
			marshalUint(w, name, 64, conv)
		case "uint32":
			marshalUint(w, name, 32, conv)
		case "uint16":
			marshalUint(w, name, 16, conv)
		case "uint8":
			marshalUint(w, name, 8, conv)
		case "string":
			marshalString(w, name, conv)
		case "[]byte":
			fpl(w, "w.WriteBytes(%s)", name)
		default:
			marshalMarshaler(w, name)
		}
	case gengo.ArrayKind:
		switch typ.Get("array_prefix").(string) {
		case "length":
			marshalLengthArray(w, name, typ)
		case "size":
			marshalSizeArray(w, name, typ)
		}
	default:
		fpl(w, "// value %s %v", name, typ.Kind)
	}
}

func unmarshalValue(w io.Writer, name string, typ *gengo.Type, declType string) {
	switch typ.Kind {
	case gengo.IdentKind:
		conv := ""
		if typ.Ident != declType {
			conv = declType
		}
		switch typ.Ident {
		case "int64":
			unmarshalInt(w, name, 64, conv)
		case "int32":
			unmarshalInt(w, name, 32, conv)
		case "int16":
			unmarshalInt(w, name, 16, conv)
		case "int8":
			unmarshalInt(w, name, 8, conv)
		case "uint64":
			unmarshalUint(w, name, 64, conv)
		case "uint32":
			unmarshalUint(w, name, 32, conv)
		case "uint16":
			unmarshalUint(w, name, 16, conv)
		case "uint8":
			unmarshalUint(w, name, 8, conv)
		case "string":
			unmarshalString(w, name, conv)
		case "[]byte":
			fpl(w, "%s = r.ReadBytes()", name)
		default:
			unmarshalUnmarshaler(w, name, typ.Ident)
		}
	case gengo.ArrayKind:
		switch typ.Get("array_prefix").(string) {
		case "length":
			unmarshalLengthArray(w, name, typ)
		case "size":
			unmarshalSizeArray(w, name, typ)
		}
	default:
		fpl(w, "// value %s %v", name, typ.Kind)
	}
}

func marshalLengthArray(w io.Writer, name string, typ *gengo.Type) {
	marshalInt(w, fmt.Sprintf("int32(len(%s))", name), 32, "")
	fpl(w, "for i := range %s {", name)
	switch typ.Ident {
	case "int8", "int16", "int32", "int64", "string":
		marshalValue(w, name+"[i]", &gengo.Type{Kind: gengo.IdentKind, Ident: typ.Ident}, typ.Ident)
	default:
		marshalMarshaler(w, name+"[i]")
	}
	fpl(w, "}")
}

func unmarshalLengthArray(w io.Writer, name string, typ *gengo.Type) {
	fpl(w, "%s = make([]%s, int(r.ReadInt32()))", name, typ.Ident)
	fpl(w, "for i := range %s {", name)
	switch typ.Ident {
	case "int8", "int16", "int32", "int64", "string":
		unmarshalValue(w, name+"[i]", &gengo.Type{Kind: gengo.IdentKind, Ident: typ.Ident}, typ.Ident)
	default:
		unmarshalUnmarshaler(w, name+"[i]", typ.Ident)
	}
	fpl(w, "}")
}

func marshalSizeArray(w io.Writer, name string, typ *gengo.Type) {
	fpl(w, "offset := len(w.B)")
	fpl(w, "w.WriteInt32(0)")
	fpl(w, "start := len(w.B)")
	fpl(w, "for i := range %s {", name)
	switch typ.Ident {
	case "int8", "int16", "int32", "int64", "string":
		marshalValue(w, name+"[i]", &gengo.Type{Kind: gengo.IdentKind, Ident: typ.Ident}, typ.Ident)
	default:
		marshalMarshaler(w, name+"[i]")
	}
	fpl(w, "}")
	fpl(w, "w.SetInt32(offset, int32(len(w.B)-start))")
}

func unmarshalSizeArray(w io.Writer, name string, typ *gengo.Type) {
	fpl(w, "size := int(r.ReadInt32())")
	fpl(w, "start := r.Offset()")
	fpl(w, "for r.Offset()-start < size {")
	fpl(w, "var m %s", typ.Ident)
	switch typ.Ident {
	case "int8", "int16", "int32", "int64", "string":
		unmarshalValue(w, "m", &gengo.Type{Kind: gengo.IdentKind, Ident: typ.Ident}, typ.Ident)
	default:
		unmarshalUnmarshaler(w, "m", typ.Ident)
	}
	// NOTE: server optimization, ignore err for sized array element and quit silently
	fpl(w, "if r.Err() != nil {")
	fpl(w, "r.SetErr(nil)")
	fpl(w, "return")
	fpl(w, "}")
	fpl(w, "*t = append(*t, m)")
	fpl(w, "}")
}

func marshalMarshaler(w io.Writer, marshaler string) {
	fpl(w, "%s.Marshal(w)", marshaler)
}

func unmarshalUnmarshaler(w io.Writer, unmarshaler, typ string) {
	fpl(w, "%s.Unmarshal(r)", unmarshaler)
}

func marshalInt(w io.Writer, name string, bit int, convType string) {
	if convType != "" {
		fpl(w, "w.WriteInt%d(int%d(%s))", bit, bit, name)
		return
	}
	fpl(w, "w.WriteInt%d(%s)", bit, name)
}

func unmarshalInt(w io.Writer, name string, bit int, convType string) {
	if convType != "" {
		fpl(w, "%s=%s(r.ReadInt%d())", name, convType, bit)
		return
	}
	fpl(w, "%s=r.ReadInt%d()", name, bit)
}

func marshalUint(w io.Writer, name string, bit int, convType string) {
	if convType != "" {
		fpl(w, "w.WriteUint%d(uint%d(%s))", bit, bit, name)
		return
	}
	fpl(w, "w.WriteUint%d(%s)", bit, name)
}

func unmarshalUint(w io.Writer, name string, bit int, convType string) {
	if convType != "" {
		fpl(w, "%s=%s(r.ReadUint%d())", name, convType, bit)
		return
	}
	fpl(w, "%s=r.ReadUint%d()", name, bit)
}

func marshalString(w io.Writer, name string, convType string) {
	if convType != "" {
		fpl(w, "w.WriteString(string(%s))", name)
		return
	}
	fpl(w, "w.WriteString(%s)", name)
}

func unmarshalString(w io.Writer, name string, convType string) {
	if convType != "" {
		fpl(w, "%s = %s(r.ReadString())", name, convType)
		return
	}
	fpl(w, name+" = r.ReadString()")
}
