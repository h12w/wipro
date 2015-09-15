package wipro

import (
	"fmt"
	"io"
	"os"

	"h12.me/gengo"
)

func FromBNFToGoFuncs() {
	w := os.Stdout
	goFile := FromBNFToGoFile()
	fpl(w, "package proto")
	fpl(w, "import (")
	fpl(w, `"hash/crc32"`)
	fpl(w, ")")
	for _, decl := range goFile.TypeDecls {
		genMarshalFunc(w, decl)
		fpl(w, "")
		genUnmarshalFunc(w, decl)
		fpl(w, "")
	}
}

func genMarshalFunc(w io.Writer, decl *gengo.TypeDecl) {
	t := &decl.Type
	if t.Kind == gengo.IdentKind && t.Ident == "T" {
		return
	}
	fpl(w, "func (t *%s) Marshal(w *Writer) {", decl.Name)
	switch t.Kind {
	case gengo.StructKind:
		if f0 := t.Fields[0]; len(t.Fields) == 2 &&
			(f0.Name == "Size" || f0.Name == "CRC") {
			fpl(w, "offset := len(w.B)")
			marshalField(w, t.Fields[0])
			fpl(w, "start := len(w.B)")
			marshalField(w, t.Fields[1])
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
	case gengo.ArrayKind:
		marshalValue(w, "(*t)", t)
	case gengo.IdentKind:
		marshalValue(w, "(*t)", t)
	default:
		fpl(w, "// type %s, %v", decl.Name, decl.Type)
	}
	fpl(w, "}")
}

func genUnmarshalFunc(w io.Writer, decl *gengo.TypeDecl) {
	t := &decl.Type
	if t.Kind == gengo.IdentKind && t.Ident == "T" {
		return
	}
	fpl(w, "func (t *%s) Unmarshal(r *Reader) {", decl.Name)
	switch t.Kind {
	case gengo.StructKind:
		if f0 := t.Fields[0]; len(t.Fields) == 2 &&
			(f0.Name == "Size" || f0.Name == "CRC") {
			unmarshalField(w, t.Fields[0])
			fpl(w, "start := r.Offset")
			unmarshalField(w, t.Fields[1])
			switch f0.Name {
			case "Size":
				fpl(w, "if r.Err == nil && int(t.Size) != r.Offset-start {")
				fpl(w, `r.Err = newError("size mismatch, expect %%d, got %%d", int(t.Size), r.Offset-start)`)
				fpl(w, "}")
			case "CRC":
				fpl(w, "if r.Err == nil && t.CRC != crc32.ChecksumIEEE(r.B[start:r.Offset]) {")
				fpl(w, `r.Err = newError("CRC mismatch")`)
				fpl(w, "}")
			}
		} else {
			for _, f := range t.Fields {
				unmarshalField(w, f)
			}
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
		fName = "t." + f.Type.Ident
	}
	marshalValue(w, fName, &f.Type)
}

func unmarshalField(w io.Writer, f *gengo.Field) {
	fName := "t." + f.Name
	if f.Name == "" {
		fName = "t." + f.Type.Ident
	}
	unmarshalValue(w, fName, &f.Type, f.Type.Ident)
}

func marshalValue(w io.Writer, name string, typ *gengo.Type) {
	switch typ.Kind {
	case gengo.IdentKind:
		switch typ.Ident {
		case "int64":
			marshalInt(w, name, 64)
		case "int32":
			marshalInt(w, name, 32)
		case "int16":
			marshalInt(w, name, 16)
		case "int8":
			marshalInt(w, name, 8)
		case "uint64":
			marshalUint(w, name, 64)
		case "uint32":
			marshalUint(w, name, 32)
		case "uint16":
			marshalUint(w, name, 16)
		case "uint8":
			marshalUint(w, name, 8)
		case "string":
			fpl(w, "w.WriteString(string(%s))", name)
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
		switch typ.Ident {
		case "int64":
			unmarshalInt(w, name, "b", 64)
		case "int32":
			unmarshalInt(w, name, "b", 32)
		case "int16":
			unmarshalInt(w, name, "b", 16)
		case "int8":
			unmarshalInt(w, name, "b", 8)
		case "uint64":
			unmarshalUint(w, name, "b", 64)
		case "uint32":
			unmarshalUint(w, name, "b", 32)
		case "uint16":
			unmarshalUint(w, name, "b", 16)
		case "uint8":
			unmarshalUint(w, name, "b", 8)
		case "string":
			fpl(w, "%s = %s(r.ReadString())", name, declType)
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
	marshalInt(w, fmt.Sprintf("int32(len(%s))", name), 32)
	fpl(w, "for i := range %s {", name)
	switch typ.Ident {
	case "int8", "int16", "int32", "int64", "string":
		marshalValue(w, name+"[i]", &gengo.Type{Kind: gengo.IdentKind, Ident: typ.Ident})
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
		marshalValue(w, name+"[i]", &gengo.Type{Kind: gengo.IdentKind, Ident: typ.Ident})
	default:
		marshalMarshaler(w, name+"[i]")
	}
	fpl(w, "}")
	fpl(w, "w.SetInt32(offset, int32(len(w.B)-start))")
}

func unmarshalSizeArray(w io.Writer, name string, typ *gengo.Type) {
	fpl(w, "size := int(r.ReadInt32())")
	fpl(w, "start := r.Offset")
	fpl(w, "for r.Offset-start < size {")
	fpl(w, "var m %s", typ.Ident)
	switch typ.Ident {
	case "int8", "int16", "int32", "int64", "string":
		unmarshalValue(w, "m", &gengo.Type{Kind: gengo.IdentKind, Ident: typ.Ident}, typ.Ident)
	default:
		unmarshalUnmarshaler(w, "m", typ.Ident)
	}
	fpl(w, "if r.Err != nil {")
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

func marshalInt(w io.Writer, name string, bit int) {
	fpl(w, "w.WriteInt%d(%s)", bit, name)
}

func unmarshalInt(w io.Writer, name string, bufName string, bit int) {
	fpl(w, "%s=r.ReadInt%d()", name, bit)
}

func marshalUint(w io.Writer, name string, bit int) {
	fpl(w, "w.WriteUint%d(%s)", bit, name)
}

func unmarshalUint(w io.Writer, name string, bufName string, bit int) {
	fpl(w, "%s=r.ReadUint%d()", name, bit)
}
