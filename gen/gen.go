package gen

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"h12.io/gengo"
)

func (bnf BNF) GoTypes() GoTypes {
	declMap := make(map[string]*Decl)
	for _, decl := range bnf {
		if foundDecl, ok := declMap[decl.Name]; ok && !reflect.DeepEqual(foundDecl, decl) {
			log.Fatalf("conflict name %s:\n%#v\n%#v", decl.Name, foundDecl, decl)
		}
		declMap[decl.Name] = decl
	}
	types := genGoTypes(bnf, declMap)
	types.File.Imports = []*gengo.Import{
		{Path: "h12.io/wipro"},
	}
	return types
}

func genGoTypes(decls BNF, declMap map[string]*Decl) GoTypes {
	var goDecls []*gengo.TypeDecl
	for _, decl := range decls {
		if !decl.Type.simple() {
			goDecls = append(goDecls, decl.GenDecl(declMap))
		}
	}
	return GoTypes{&gengo.File{
		TypeDecls: goDecls,
	}}
}

func (n *Node) simple() bool {
	if len(n.Child) != 1 || n.Child[0].NodeType != LeafNode {
		return false
	}
	switch n.Child[0].Value {
	case "string", "bytes", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64":
		return true
	}
	return false
}

func (n *Node) GenType(m map[string]*Decl) *gengo.Type {
	switch n.NodeType {
	case LeafNode:
		if decl, ok := m[n.Value]; ok && decl.Type.simple() {
			return decl.Type.GenType(m)
		} else {
			return &gengo.Type{Kind: gengo.IdentKind, Ident: goType(n.Value)}
		}
	case SeqNode:
		if len(n.Child) == 1 && n.Child[0].NodeType == LeafNode {
			return n.Child[0].GenType(m)
		} else {
			var fields []*gengo.Field
			for _, c := range n.Child {
				fields = append(fields, c.GenField(m))
			}
			return &gengo.Type{
				Kind:   gengo.StructKind,
				Fields: fields,
			}
		}
	case LengthArrayNode, SizeArrayNode:
		var t *gengo.Type
		if len(n.Child) == 1 {
			t = n.Child[0].GenType(m)
		} else {
			t = (&Node{NodeType: SeqNode, Child: n.Child}).GenType(m)
		}
		t.Kind = gengo.ArrayKind
		switch n.NodeType {
		case LengthArrayNode:
			t.Set("array_prefix", "length")
		case SizeArrayNode:
			t.Set("array_prefix", "size")
		}
		return t
	case OrNode:
		return &gengo.Type{
			Kind:  gengo.IdentKind,
			Ident: "wipro.M",
		}
	default:
		return &gengo.Type{
			Kind:  gengo.IdentKind,
			Ident: "-",
		}
	}
}

func (d *Decl) GenDecl(m map[string]*Decl) *gengo.TypeDecl {
	ut := d.Type.Value
	for ut != "" {
		if t, ok := m[ut]; ok && t.Type.Value != "" {
			ut = t.Type.Value
		} else {
			break
		}
	}
	return &gengo.TypeDecl{
		Name: goName(d.Name),
		Type: *d.Type.GenType(m),
	}
}

func (d *Decl) typeName() string {
	typ := ""
	if d.Type.simple() {
		typ = d.Type.Child[0].Value
	} else {
		typ = d.Name
	}
	return goType(typ)
}

func (n *Node) GenField(m map[string]*Decl) *gengo.Field {
	name := n.Value
	decl, _ := m[name]
	if name == "" && n.NodeType == LengthArrayNode && len(n.Child) == 1 {
		name = n.Child[0].Value + "s"
	}
	if decl != nil && n.NodeType == LengthArrayNode {
		goType := decl.Type.GenType(m)
		if n.NodeType == LengthArrayNode {
			goType.Kind = gengo.ArrayKind
		}
		return &gengo.Field{
			Name: goName(name),
			Type: *goType,
		}
	}
	return &gengo.Field{
		Name: goName(name),
		Type: *n.GenType(m),
	}
}

func fp(w io.Writer, format string, v ...interface{}) {
	fmt.Fprintf(w, format, v...)
}

func fpl(w io.Writer, format string, v ...interface{}) {
	fmt.Fprintf(w, format, v...)
	fmt.Fprintln(w)
}

func goName(s string) string {
	if strings.HasSuffix(s, "Id") {
		return strings.TrimSuffix(s, "Id") + "ID"
	}
	s = strings.Replace(s, "Crc", "CRC", -1)
	s = strings.Replace(s, "Api", "API", -1)
	s = strings.Replace(s, "Isr", "ISR", -1)
	return s
}

func goType(s string) string {
	switch s {
	case "bytes":
		return "[]byte"
	}
	return goName(s)
}
