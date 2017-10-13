package msgplens

import (
	"encoding/binary"
	"fmt"
)

// Copied directly from http://github.com/tinylib/msgp/msgp

// Copyright (c) 2014 Philip Hofer
// Portions Copyright (c) 2009 The Go Authors (license at http://golang.org) where indicated
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

var big = binary.BigEndian

const (
	// 0XXXXXXX
	Fixint uint8 = 0x00

	// 111XXXXX
	FixintNeg uint8 = 0xe0

	// 1000XXXX
	Fixmap uint8 = 0x80

	// 1001XXXX
	Fixarray uint8 = 0x90

	// 101XXXXX
	Fixstr uint8 = 0xa0

	Nil      uint8 = 0xc0
	False    uint8 = 0xc2
	True     uint8 = 0xc3
	Bin8     uint8 = 0xc4
	Bin16    uint8 = 0xc5
	Bin32    uint8 = 0xc6
	Ext8     uint8 = 0xc7
	Ext16    uint8 = 0xc8
	Ext32    uint8 = 0xc9
	Float32  uint8 = 0xca
	Float64  uint8 = 0xcb
	Uint8    uint8 = 0xcc
	Uint16   uint8 = 0xcd
	Uint32   uint8 = 0xce
	Uint64   uint8 = 0xcf
	Int8     uint8 = 0xd0
	Int16    uint8 = 0xd1
	Int32    uint8 = 0xd2
	Int64    uint8 = 0xd3
	Fixext1  uint8 = 0xd4
	Fixext2  uint8 = 0xd5
	Fixext4  uint8 = 0xd6
	Fixext8  uint8 = 0xd7
	Fixext16 uint8 = 0xd8
	Str8     uint8 = 0xd9
	Str16    uint8 = 0xda
	Str32    uint8 = 0xdb
	Array16  uint8 = 0xdc
	Array32  uint8 = 0xdd
	Map16    uint8 = 0xde
	Map32    uint8 = 0xdf
)

// size mode
// if positive, # elements for composites
type varmode int8

const (
	constsize varmode = 0  // constant size (size bytes + uint8(varmode) objects)
	extra8            = -1 // has uint8(p[1]) extra bytes
	extra16           = -2 // has be16(p[1:]) extra bytes
	extra32           = -3 // has be32(p[1:]) extra bytes
	map16v            = -4 // use map16
	map32v            = -5 // use map32
	array16v          = -6 // use array16
	array32v          = -7 // use array32
)

func getType(v byte) Type {
	return sizes[v].typ
}

var sizes = [256]bytespec{
	Nil:      {size: 1, extra: constsize, typ: NilType},
	False:    {size: 1, extra: constsize, typ: BoolType},
	True:     {size: 1, extra: constsize, typ: BoolType},
	Bin8:     {size: 2, extra: extra8, typ: BinType},
	Bin16:    {size: 3, extra: extra16, typ: BinType},
	Bin32:    {size: 5, extra: extra32, typ: BinType},
	Ext8:     {size: 3, extra: extra8, typ: ExtensionType},
	Ext16:    {size: 4, extra: extra16, typ: ExtensionType},
	Ext32:    {size: 6, extra: extra32, typ: ExtensionType},
	Float32:  {size: 5, extra: constsize, typ: Float32Type},
	Float64:  {size: 9, extra: constsize, typ: Float64Type},
	Uint8:    {size: 2, extra: constsize, typ: UintType},
	Uint16:   {size: 3, extra: constsize, typ: UintType},
	Uint32:   {size: 5, extra: constsize, typ: UintType},
	Uint64:   {size: 9, extra: constsize, typ: UintType},
	Int8:     {size: 2, extra: constsize, typ: IntType},
	Int16:    {size: 3, extra: constsize, typ: IntType},
	Int32:    {size: 5, extra: constsize, typ: IntType},
	Int64:    {size: 9, extra: constsize, typ: IntType},
	Fixext1:  {size: 3, extra: constsize, typ: ExtensionType},
	Fixext2:  {size: 4, extra: constsize, typ: ExtensionType},
	Fixext4:  {size: 6, extra: constsize, typ: ExtensionType},
	Fixext8:  {size: 10, extra: constsize, typ: ExtensionType},
	Fixext16: {size: 18, extra: constsize, typ: ExtensionType},
	Str8:     {size: 2, extra: extra8, typ: StrType},
	Str16:    {size: 3, extra: extra16, typ: StrType},
	Str32:    {size: 5, extra: extra32, typ: StrType},
	Array16:  {size: 3, extra: array16v, typ: ArrayType},
	Array32:  {size: 5, extra: array32v, typ: ArrayType},
	Map16:    {size: 3, extra: map16v, typ: MapType},
	Map32:    {size: 5, extra: map32v, typ: MapType},
}

func init() {
	// set up fixed fields

	// fixint
	for i := Fixint; i < 0x80; i++ {
		sizes[i] = bytespec{size: 1, extra: constsize, typ: IntType}
	}

	// nfixint
	for i := uint16(FixintNeg); i < 0x100; i++ {
		sizes[uint8(i)] = bytespec{size: 1, extra: constsize, typ: IntType}
	}

	// fixstr gets constsize,
	// since the prefix yields the size
	for i := Fixstr; i < 0xc0; i++ {
		sizes[i] = bytespec{size: 1 + rfixstr(i), extra: constsize, typ: StrType}
	}

	// fixmap
	for i := Fixmap; i < 0x90; i++ {
		sizes[i] = bytespec{size: 1, extra: varmode(2 * rfixmap(i)), typ: MapType}
	}

	// fixarray
	for i := Fixarray; i < 0xa0; i++ {
		sizes[i] = bytespec{size: 1, extra: varmode(rfixarray(i)), typ: ArrayType}
	}
}

const last4 = 0x0f
const first4 = 0xf0
const last5 = 0x1f
const first3 = 0xe0
const last7 = 0x7f

func rfixstr(b byte) uint8   { return b & last5 }
func rfixarray(b byte) uint8 { return b & last4 }
func rfixmap(b byte) uint8   { return b & last4 }

// a valid bytespsec has
// non-zero 'size' and
// non-zero 'typ'
type bytespec struct {
	size  uint8   // prefix size information
	extra varmode // extra size information
	typ   Type    // type
	_     byte    // makes bytespec 4 bytes (yes, this matters)
}

// Type is a MessagePack wire type,
// including this package's built-in
// extension types.
type Type byte

// MessagePack Types
//
// The zero value of Type
// is InvalidType.
const (
	InvalidType Type = iota

	// MessagePack built-in types

	StrType
	BinType
	MapType
	ArrayType
	Float64Type
	Float32Type
	BoolType
	IntType
	UintType
	NilType
	ExtensionType

	_maxtype
)

// String implements fmt.Stringer
func (t Type) String() string {
	switch t {
	case StrType:
		return "str"
	case BinType:
		return "bin"
	case MapType:
		return "map"
	case ArrayType:
		return "array"
	case Float64Type:
		return "float64"
	case Float32Type:
		return "float32"
	case BoolType:
		return "bool"
	case UintType:
		return "uint"
	case IntType:
		return "int"
	case ExtensionType:
		return "ext"
	case NilType:
		return "nil"
	default:
		return "<invalid>"
	}
}

// returns (skip N bytes, skip M objects, error)
func getSize(b []byte) (uintptr, uintptr, error) {
	l := len(b)
	if l == 0 {
		return 0, 0, fmt.Errorf("short read")
	}
	lead := b[0]
	spec := &sizes[lead] // get type information
	size, mode := spec.size, spec.extra
	if size == 0 {
		return 0, 0, fmt.Errorf("invalid prefix %v", lead)
	}
	if mode >= 0 { // fixed composites
		return uintptr(size), uintptr(mode), nil
	}
	if l < int(size) {
		return 0, 0, fmt.Errorf("short read")
	}
	switch mode {
	case extra8:
		return uintptr(size) + uintptr(b[1]), 0, nil
	case extra16:
		return uintptr(size) + uintptr(big.Uint16(b[1:])), 0, nil
	case extra32:
		return uintptr(size) + uintptr(big.Uint32(b[1:])), 0, nil
	case map16v:
		return uintptr(size), 2 * uintptr(big.Uint16(b[1:])), nil
	case map32v:
		return uintptr(size), 2 * uintptr(big.Uint32(b[1:])), nil
	case array16v:
		return uintptr(size), uintptr(big.Uint16(b[1:])), nil
	case array32v:
		return uintptr(size), uintptr(big.Uint32(b[1:])), nil
	default:
		return 0, 0, fmt.Errorf("fatal")
	}
}
