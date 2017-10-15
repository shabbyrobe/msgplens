package msgplens

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
)

var byteOrder = binary.BigEndian

type Node interface {
	Msgpack(into *bytes.Buffer) error
	TotalSize() int

	setCommon(prefix uint8, size int)
}

type commonNode struct {
	Prefix uint8
	Size   int
}

func (c *commonNode) setCommon(prefix uint8, size int) {
	c.Prefix = prefix
	c.Size = size
}

func (c *commonNode) TotalSize() int {
	return c.Size
}

type FloatNode struct {
	commonNode
	Bits   []byte
	Approx float64
}

func (n *FloatNode) Msgpack(into *bytes.Buffer) error {
	typ := sizes[n.Prefix].typ
	switch typ {
	case Float64Type:
		into.WriteByte(n.Prefix)
		into.Write(n.Bits)

	case Float32Type:
		into.WriteByte(n.Prefix)
		into.Write(n.Bits)

	default:
		return fmt.Errorf("unexpected number type %s", typ)
	}
	return nil
}

type IntNode struct {
	commonNode
	Bits   []byte
	Approx int64
}

func (n *IntNode) Msgpack(into *bytes.Buffer) error {
	typ := sizes[n.Prefix].typ
	switch typ {
	case IntType:
		if isfixint(n.Prefix) || isnfixint(n.Prefix) {
			into.WriteByte(n.Prefix)
		} else {
			u := byteOrder.Uint64(n.Bits)
			i := int64(u)
			var b []byte
			switch n.Prefix {
			case Int64:
				b = make([]byte, 9)
				putMint64(b, i)
				into.Write(b)
			case Int32:
				b = make([]byte, 5)
				putMint32(b, int32(i))
				into.Write(b)
			case Int16:
				b = make([]byte, 3)
				putMint16(b, int16(i))
				into.Write(b)
			case Int8:
				b = make([]byte, 2)
				putMint8(b, int8(i))
				into.Write(b)
			default:
				return fmt.Errorf("unexpected number type %s, prefix %02x", typ, n.Prefix)
			}
		}

	default:
		return fmt.Errorf("unexpected number type %s", typ)
	}
	return nil
}

type UintNode struct {
	commonNode
	Bits   []byte
	Approx uint64
}

func (n *UintNode) Msgpack(into *bytes.Buffer) error {
	typ := sizes[n.Prefix].typ
	switch typ {
	case UintType:
		if isfixint(n.Prefix) {
			into.WriteByte(n.Prefix)
		} else {
			u := byteOrder.Uint64(n.Bits)
			var b []byte
			switch n.Prefix {
			case Uint64:
				b = make([]byte, 9)
				putMuint64(b, u)
			case Uint32:
				b = make([]byte, 5)
				putMuint32(b, uint32(u))
			case Uint16:
				b = make([]byte, 3)
				putMuint16(b, uint16(u))
			case Uint8:
				b = make([]byte, 2)
				putMuint8(b, uint8(u))
			default:
				return fmt.Errorf("unexpected number type %s, prefix %02x", typ, n.Prefix)
			}
			into.Write(b)
		}

	default:
		return fmt.Errorf("unexpected number type %s", typ)
	}
	return nil
}

type BoolNode struct {
	commonNode
	Value bool
}

func (b *BoolNode) Msgpack(into *bytes.Buffer) error {
	if b.Value {
		into.WriteByte(True)
	} else {
		into.WriteByte(False)
	}
	return nil
}

type NilNode struct {
	commonNode
}

func (n *NilNode) Msgpack(into *bytes.Buffer) error {
	into.WriteByte(Nil)
	return nil
}

type ExtensionNode struct {
	commonNode
	Contents []byte
}

func (e *ExtensionNode) Msgpack(into *bytes.Buffer) error {
	into.Write(e.Contents)
	return nil
}

func (e *ExtensionNode) TotalSize() int {
	return e.Size + len(e.Contents)
}

type StrNode struct {
	commonNode
	Value string
}

func (s *StrNode) Msgpack(into *bytes.Buffer) error {
	var b [5]byte
	bs, err := writeStringHeader(s.Prefix, b[:], uint32(len(s.Value)))
	if err != nil {
		return err
	}
	into.Write(bs)
	into.WriteString(s.Value)
	return nil
}

func (s *StrNode) TotalSize() int {
	return s.Size + len(s.Value)
}

type BinNode struct {
	commonNode
	Value []byte
}

func (b *BinNode) Msgpack(into *bytes.Buffer) error {
	var bb [5]byte
	bs, err := writeBinHeader(b.Prefix, bb[:], uint32(len(b.Value)))
	if err != nil {
		return err
	}

	into.Write(bs)
	into.Write(b.Value)
	return nil
}

func (b *BinNode) TotalSize() int {
	return b.Size + len(b.Value)
}

type ArrayNode struct {
	commonNode
	Children NodeList `json:",omitempty"`
}

func (a *ArrayNode) Msgpack(into *bytes.Buffer) error {
	var bb [5]byte
	bs, err := writeArrayHeader(a.Prefix, bb[:], uint32(len(a.Children)))
	if err != nil {
		return err
	}
	into.Write(bs)

	for _, c := range a.Children {
		if err := c.Msgpack(into); err != nil {
			return err
		}
	}
	return nil
}

func (a *ArrayNode) TotalSize() int {
	sz := a.Size
	for _, c := range a.Children {
		sz += c.TotalSize()
	}
	return sz
}

type MapNode struct {
	commonNode
	Values []KeyValueNode `json:",omitempty"`
}

func (m *MapNode) Msgpack(into *bytes.Buffer) error {
	var bb [5]byte
	bs, err := writeMapHeader(m.Prefix, bb[:], uint32(len(m.Values)))
	if err != nil {
		return err
	}
	into.Write(bs)

	for _, c := range m.Values {
		if err := c.Key.Msgpack(into); err != nil {
			return err
		}
		if err := c.Value.Msgpack(into); err != nil {
			return err
		}
	}
	return nil
}

func (m *MapNode) TotalSize() int {
	sz := m.Size
	for _, c := range m.Values {
		sz += c.Key.TotalSize() + c.Value.TotalSize()
	}
	return sz
}

type KeyValueNode struct {
	Key   Node
	Value Node
}

func (k *KeyValueNode) UnmarshalJSON(in []byte) (err error) {
	var ms map[string]json.RawMessage
	if err = json.Unmarshal(in, &ms); err != nil {
		return
	}
	kb, ok := ms["Key"]
	if !ok {
		return fmt.Errorf("key not found")
	}
	if (*k).Key, err = ReprUnmarshalNode(kb); err != nil {
		return
	}
	vb, ok := ms["Value"]
	if !ok {
		return fmt.Errorf("value not found")
	}
	if (*k).Value, err = ReprUnmarshalNode(vb); err != nil {
		return
	}
	return
}

type NodeList []Node

func (n *NodeList) UnmarshalJSON(in []byte) (err error) {
	var cs []json.RawMessage
	if err = json.Unmarshal(in, &cs); err != nil {
		return
	}
	*n = make(NodeList, len(cs))
	for i, c := range cs {
		if (*n)[i], err = ReprUnmarshalNode(c); err != nil {
			return err
		}
	}
	return nil
}

func ReprUnmarshalNode(in []byte) (Node, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(in, &m); err != nil {
		return nil, err
	}

	prefix, ok := m["Prefix"]
	if !ok {
		return nil, fmt.Errorf("unknown prefix")
	}

	prefixFloat, ok := prefix.(float64)
	if !ok {
		return nil, fmt.Errorf("unknown prefix type")
	}

	prefixInt := int(prefixFloat)
	if prefixInt < 0 || prefixInt > 255 {
		return nil, fmt.Errorf("invalid prefix %d", prefixInt)
	}

	prefixByte := byte(prefixInt)
	typ := sizes[prefixByte].typ

	var node Node
	switch typ {
	case StrType:
		node = &StrNode{}
	case BinType:
		node = &BinNode{}
	case MapType:
		node = &MapNode{}
	case ArrayType:
		node = &ArrayNode{}
	case BoolType:
		node = &BoolNode{}
	case IntType:
		node = &IntNode{}
	case UintType:
		node = &UintNode{}
	case Float64Type:
		node = &FloatNode{}
	case Float32Type:
		node = &FloatNode{}
	case NilType:
		node = &NilNode{}
	case ExtensionType:
		node = &ExtensionNode{}
	default:
		return nil, fmt.Errorf("unknown type %s", typ)
	}

	if err := json.Unmarshal(in, node); err != nil {
		return nil, err
	}

	return node, nil
}

type Representer struct {
	vis   *Visitor
	root  NodeList
	nodes *NodeList
	stack []*NodeList
}

func (r *Representer) Visitor() *Visitor {
	return r.vis
}

func (r *Representer) Nodes() []Node {
	return r.root
}

func NewRepresenter() *Representer {
	r := &Representer{}
	r.nodes = &r.root

	r.vis = &Visitor{
		Int: func(ctx *LensContext, bts []byte, data int64) error {
			bits := make([]byte, 8)
			byteOrder.PutUint64(bits, uint64(data))
			*r.nodes = append(*r.nodes, &IntNode{
				commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
				Bits:       bts,
				Approx:     data})
			return nil
		},

		Uint: func(ctx *LensContext, bts []byte, data uint64) error {
			bits := make([]byte, 8)
			byteOrder.PutUint64(bits, data)
			*r.nodes = append(*r.nodes, &UintNode{
				commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
				Bits:       bts,
				Approx:     data})
			return nil
		},

		Float64: func(ctx *LensContext, bts []byte, data float64) error {
			bits := make([]byte, 8)
			byteOrder.PutUint64(bits, math.Float64bits(data))
			*r.nodes = append(*r.nodes, &FloatNode{
				commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
				Bits:       bts,
				Approx:     data})
			return nil
		},

		Float32: func(ctx *LensContext, bts []byte, data float32) error {
			bits := make([]byte, 4)
			byteOrder.PutUint32(bits, math.Float32bits(data))
			*r.nodes = append(*r.nodes, &FloatNode{
				commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
				Bits:       bts,
				Approx:     float64(data)})
			return nil
		},

		Str: func(ctx *LensContext, bts []byte, str string) error {
			*r.nodes = append(*r.nodes, &StrNode{
				commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
				Value:      str})
			return nil
		},

		Bin: func(ctx *LensContext, bts []byte, data []byte) error {
			*r.nodes = append(*r.nodes, &BinNode{
				commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
				Value:      data})
			return nil
		},

		Bool: func(ctx *LensContext, bts []byte, data bool) error {
			*r.nodes = append(*r.nodes, &BoolNode{
				commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
				Value:      data})
			return nil
		},

		Nil: func(ctx *LensContext, prefix byte) error {
			*r.nodes = append(*r.nodes, &NilNode{
				commonNode: commonNode{Prefix: Nil, Size: 1}})
			return nil
		},

		Extension: func(ctx *LensContext, bts []byte) error {
			*r.nodes = append(*r.nodes, &ExtensionNode{
				commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
				Contents:   bts})
			return nil
		},

		EnterArray: func(ctx *LensContext, prefix byte, objects int) error {
			an := &ArrayNode{
				commonNode: commonNode{Prefix: prefix, Size: int(sizes[prefix].size)},
			}
			*r.nodes = append(*r.nodes, an)
			r.stack = append(r.stack, r.nodes)
			r.nodes = &an.Children
			return nil
		},

		LeaveArray: func(ctx *LensContext, prefix byte, cnt int, bts []byte) error {
			r.stack, r.nodes = r.stack[0:len(r.stack)-1], r.stack[len(r.stack)-1]
			return nil
		},

		EnterMap: func(ctx *LensContext, prefix byte, cnt int) error {
			m := NodeList{}
			r.stack = append(r.stack, r.nodes)
			r.nodes = &m
			return nil
		},

		LeaveMap: func(ctx *LensContext, prefix byte, cnt int, bts []byte) error {
			mn := &MapNode{
				commonNode: commonNode{Prefix: prefix, Size: int(sizes[prefix].size)},
			}
			ln := len(*r.nodes)
			for i := 0; i < ln; i += 2 {
				mn.Values = append(mn.Values, KeyValueNode{
					Key:   (*r.nodes)[i],
					Value: (*r.nodes)[i+1],
				})
			}
			r.stack, r.nodes = r.stack[0:len(r.stack)-1], r.stack[len(r.stack)-1]
			*r.nodes = append(*r.nodes, mn)
			return nil
		},

		End: func(ctx *LensContext, left []byte) error {
			*r.nodes = append(*r.nodes, &BinNode{
				commonNode: commonNode{Prefix: Bin32, Size: len(left)},
				Value:      left})
			return nil
		},
	}
	return r
}
