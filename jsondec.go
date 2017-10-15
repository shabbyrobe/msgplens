package msgplens

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// UnmarshalJSON unmarshals a lossy JSON representation of a msgpack object
// into a Node.
func UnmarshalJSON(b []byte, extra bool) (Node, error) {
	var intf interface{}
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err := decoder.Decode(&intf); err != nil {
		return nil, err
	}
	if decoder.More() && !extra {
		return nil, fmt.Errorf("extra bytes after JSON object")
	}
	node, err := jsonIntfToNode(intf)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func jsonIntfToNode(intf interface{}) (Node, error) {
	//	bool, for JSON booleans
	//	float64, for JSON numbers
	//	string, for JSON strings
	//	[]interface{}, for JSON arrays
	//	map[string]interface{}, for JSON objects
	//	nil for JSON null

	switch v := intf.(type) {
	case bool:
		if v == true {
			return &BoolNode{commonNode: commonNode{Prefix: True, Size: int(sizes[True].size)}, Value: v}, nil
		} else {
			return &BoolNode{commonNode: commonNode{Prefix: False, Size: int(sizes[True].size)}, Value: v}, nil
		}

	case json.Number:
		if strings.ContainsAny(v.String(), ".eE") {
			n, err := v.Float64()
			if err != nil {
				return nil, err
			}
			bits := make([]byte, 8)
			byteOrder.PutUint64(bits, math.Float64bits(n))
			return &FloatNode{
				commonNode: commonNode{Prefix: Float64, Size: int(sizes[Float64].size)},
				Bits:       bits,
				Approx:     n}, nil

		} else {
			n, err := v.Int64()
			if err != nil {
				return nil, err
			}
			bits := make([]byte, 8)

			var prefix byte
			var node Node

			switch {
			case n >= 0 && n < 128:
				prefix = wfixint(byte(n))
				node = &IntNode{Approx: n, Bits: bits}
			case n < 0 && n >= -32 && isnfixint(byte(n)):
				prefix = wnfixint(int8(n))
				node = &IntNode{Approx: n, Bits: bits}
			case n >= 128 && n < math.MaxUint8:
				prefix = Uint8
				node = &UintNode{Approx: uint64(n), Bits: bits}
			case n >= math.MaxUint8 && n < math.MaxUint16:
				prefix = Uint16
				node = &UintNode{Approx: uint64(n), Bits: bits}
			case n >= math.MaxUint16 && n < math.MaxUint32:
				prefix = Uint32
				node = &UintNode{Approx: uint64(n), Bits: bits}
			case n < 0 && n >= math.MinInt8:
				prefix = Int8
				node = &IntNode{Approx: n, Bits: bits}
			case n < math.MinInt8 && n >= math.MinInt16:
				prefix = Int16
				node = &IntNode{Approx: n, Bits: bits}
			case n < math.MinInt16 && n >= math.MinInt32:
				prefix = Int16
				node = &IntNode{Approx: n, Bits: bits}
			default:
				prefix = Int64
				node = &IntNode{Approx: n, Bits: bits}
			}
			byteOrder.PutUint64(bits, uint64(n))
			node.setCommon(prefix, int(sizes[prefix].size))
			return node, nil
		}

	case string:
		var prefix byte
		sz := len(v)
		switch {
		case sz <= 31:
			prefix = wfixstr(uint8(sz))
		case sz <= math.MaxUint8:
			prefix = Str8
		case sz <= math.MaxUint16:
			prefix = Str16
		default:
			prefix = Str32
		}
		return &StrNode{
			commonNode: commonNode{Prefix: prefix, Size: int(sizes[Int64].size)},
			Value:      v}, nil

	case []interface{}:
		n := &ArrayNode{}
		for _, i := range v {
			cn, err := jsonIntfToNode(i)
			if err != nil {
				return nil, err
			}
			n.Children = append(n.Children, cn)
		}
		sz := len(n.Children)
		n.commonNode.Size = sz
		switch {
		case sz <= 15:
			n.commonNode.Prefix = wfixarray(uint8(sz))
		case sz <= math.MaxUint16:
			n.commonNode.Prefix = Array16
		default:
			n.commonNode.Prefix = Array32
		}
		return n, nil

	case map[string]interface{}:
		vlen := len(v)
		n := &MapNode{
			Values: make([]KeyValueNode, vlen),
		}
		n.commonNode.Size = vlen

		keys := make([]string, vlen)
		i := 0
		for k := range v {
			keys[i] = k
			i++
		}
		sort.Strings(keys)

		i = 0
		for _, k := range keys {
			ck, err := jsonIntfToNode(k)
			if err != nil {
				return nil, err
			}
			cv, err := jsonIntfToNode(v[k])
			if err != nil {
				return nil, err
			}
			n.Values[i] = KeyValueNode{Key: ck, Value: cv}
			i++
		}

		sz := len(n.Values) * 2
		switch {
		case sz <= 15:
			n.commonNode.Prefix = wfixmap(uint8(sz))
		case sz <= math.MaxUint16:
			n.commonNode.Prefix = Map16
		default:
			n.commonNode.Prefix = Map32
		}
		return n, nil

	default:
		return nil, fmt.Errorf("unknown type %T", v)
	}
}
