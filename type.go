package wipro

type NodeType int

const (
	LeafNode NodeType = iota
	SeqNode
	OrNode
	LengthArrayNode
	SizeArrayNode
)

func (t NodeType) MarshalText() ([]byte, error) {
	switch t {
	case LeafNode:
		return []byte(""), nil
	case SeqNode:
		return []byte(" "), nil
	case OrNode:
		return []byte("|"), nil
	case LengthArrayNode:
		return []byte("*"), nil
	}
	return nil, nil
}

type Node struct {
	NodeType NodeType `json:"node_type,omitempty"`
	Value    string   `json:"value,omitempty"`
	Child    []*Node  `json:"child,omitempty"`
}

type Decl struct {
	Name string `json:"name"`
	Type *Node  `json:"type"`
}
