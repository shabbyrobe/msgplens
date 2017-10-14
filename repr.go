package msgplens

import (
	"encoding/binary"
	"fmt"
	"math"
)

type Node interface {
	AsMap() map[string]interface{}
	FromMap(m map[string]interface{}) error
}

type commonNode struct {
	Prefix uint8
	Size   int
}

func (c commonNode) AsMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["Size"] = c.Size
	m["Prefix"] = c.Prefix
	return m
}

func (c *commonNode) FromMap(m map[string]interface{}) (err error) {
	if c.Prefix, err = mapEatUint8(m, "Prefix", true); err != nil {
		return
	}
	c.Size, err = mapEatInt(m, "Size", true)
	return
}

type ArrayNode struct {
	commonNode
	Children []Node
}

func (a *ArrayNode) AsMap() map[string]interface{} {
	m := a.commonNode.AsMap()
	if a.Children == nil {
		m["Children"] = []Node{}
	} else {
		m["Children"] = a.Children
	}
	return m
}

type NumberNode struct {
	commonNode
	Bits []byte
}

func (i *NumberNode) AsMap() map[string]interface{} {
	m := i.commonNode.AsMap()
	m["Bits"] = i.Bits
	return m
}

type BoolNode struct {
	commonNode
	Value bool
}

func (b *BoolNode) AsMap() map[string]interface{} {
	m := b.commonNode.AsMap()
	m["Value"] = b.Value
	return m
}

type NilNode struct {
	commonNode
}

type ExtensionNode struct {
	commonNode
	Contents []byte
}

func (e *ExtensionNode) AsMap() map[string]interface{} {
	m := e.commonNode.AsMap()
	m["Value"] = e.Contents
	return m
}

type StrNode struct {
	commonNode
	Value string
}

func (s *StrNode) AsMap() map[string]interface{} {
	m := s.commonNode.AsMap()
	m["Value"] = s.Value
	return m
}

type BinNode struct {
	commonNode
	Value []byte
}

func (b *BinNode) AsMap() map[string]interface{} {
	m := b.commonNode.AsMap()
	m["Value"] = b.Value
	return m
}

type KeyValueNode struct {
	Key   Node
	Value Node
}

type MapNode struct {
	commonNode
	Values []KeyValueNode
}

func (n *MapNode) AsMap() map[string]interface{} {
	m := n.commonNode.AsMap()
	m["Values"] = n.Values
	return m
}

func FromMap(m map[string]interface{}) (Node, error) {
	prefix, ok := m["Prefix"]
	if !ok {
		return nil, fmt.Errorf("unknown prefix")
	}

	prefixInt, ok := prefix.(int64)
	if prefixInt < 0 || prefixInt > 255 {
		return nil, fmt.Errorf("invalid prefix %d", prefixInt)
	}

	prefixByte := byte(prefixInt)
	typ := sizes[prefixByte].typ

	var node Node
	switch typ {
	case StrType:
		node = &StrNode{commonNode: commonNode{Prefix: prefixByte}}
	case BinType:
		node = &BinNode{commonNode: commonNode{Prefix: prefixByte}}
	case MapType:
		node = &MapNode{commonNode: commonNode{Prefix: prefixByte}}
	case ArrayType:
		node = &ArrayNode{commonNode: commonNode{Prefix: prefixByte}}
	case BoolType:
		node = &BoolNode{commonNode: commonNode{Prefix: prefixByte}}
	case IntType:
		node = &NumberNode{commonNode: commonNode{Prefix: prefixByte}}
	case UintType:
		node = &NumberNode{commonNode: commonNode{Prefix: prefixByte}}
	case Float64Type:
		node = &NumberNode{commonNode: commonNode{Prefix: prefixByte}}
	case Float32Type:
		node = &NumberNode{commonNode: commonNode{Prefix: prefixByte}}
	case NilType:
		node = &NilNode{commonNode: commonNode{Prefix: prefixByte}}
	case ExtensionType:
		node = &ExtensionNode{commonNode: commonNode{Prefix: prefixByte}}
	default:
		return nil, fmt.Errorf("unknown type %s", typ)
	}
	if err := node.FromMap(m); err != nil {
		return nil, err
	}
	return node, nil
}

type Representer struct {
	vis   *Visitor
	root  []Node
	nodes *[]Node
	stack []*[]Node
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
	numberNode := func(ctx *LensContext, bts []byte) error {
		*r.nodes = append(*r.nodes, &NumberNode{
			commonNode: commonNode{Prefix: bts[0], Size: len(bts)},
			Bits:       bts})
		return nil
	}
	r.vis = &Visitor{
		Int: func(ctx *LensContext, bts []byte, data int64) error {
			bits := make([]byte, 8)
			binary.LittleEndian.PutUint64(bits, uint64(data))
			return numberNode(ctx, bits)
		},
		Uint: func(ctx *LensContext, bts []byte, data uint64) error {
			bits := make([]byte, 8)
			binary.LittleEndian.PutUint64(bits, data)
			return numberNode(ctx, bits)
		},
		Float: func(ctx *LensContext, bts []byte, data float64) error {
			bits := make([]byte, 8)
			binary.LittleEndian.PutUint64(bits, math.Float64bits(data))
			return numberNode(ctx, bits)
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
				commonNode: commonNode{Prefix: prefix, Size: objects},
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
			m := []Node{}
			r.stack = append(r.stack, r.nodes)
			r.nodes = &m
			return nil
		},
		LeaveMap: func(ctx *LensContext, prefix byte, cnt int, bts []byte) error {
			mn := &MapNode{
				commonNode: commonNode{Prefix: prefix, Size: cnt},
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
	}
	return r
}

func mapEatInt(m map[string]interface{}, key string, req bool) (i int, err error) {
	v, ok := m[key]
	if !ok && !req {
		err = fmt.Errorf("could not find key %s", key)
		return
	}
	delete(m, key)

	i, ok = v.(int)
	if !ok {
		err = fmt.Errorf("unexpected type %T for key %s", v, key)
		return
	}
	return
}

func mapEatUint8(m map[string]interface{}, key string, req bool) (u uint8, err error) {
	var vi int
	vi, err = mapEatInt(m, key, req)
	if err != nil {
		return
	}
	if vi < 0 || vi > 255 {
		err = fmt.Errorf("value %d not between 0-255 for key %s", vi, key)
		return
	}
	u = uint8(vi)
	return
}
