package msgplens

import (
	"bytes"
	"reflect"
	"strconv"
	"unsafe"
)

// JSONEncoder exports a msgpack object as a lossy JSON equivalent.
type JSONEncoder struct {
	buf          *bytes.Buffer
	vis          *Visitor
	floatScratch []byte
}

func (j *JSONEncoder) Visitor() *Visitor {
	return j.vis
}

func NewJSONEncoder() *JSONEncoder {
	je := &JSONEncoder{}
	je.buf = &bytes.Buffer{}
	je.vis = &Visitor{
		Nil: func(ctx *LensContext, prefix byte) error { je.buf.WriteString("null"); return nil },
		Str: func(ctx *LensContext, bts []byte, str string) error { je.writeJSONString(str); return nil },
		Int: func(ctx *LensContext, bts []byte, data int64) error {
			je.buf.WriteString(strconv.FormatInt(data, 10))
			return nil
		},
		Uint: func(ctx *LensContext, bts []byte, data uint64) error {
			je.buf.WriteString(strconv.FormatUint(data, 10))
			return nil
		},
		Bin:       func(ctx *LensContext, bts []byte, data []byte) error { je.writeBin(data); return nil },
		Float64:   func(ctx *LensContext, bts []byte, data float64) error { return je.writeJSONFloat(data) },
		Float32:   func(ctx *LensContext, bts []byte, data float32) error { return je.writeJSONFloat(float64(data)) },
		Extension: func(ctx *LensContext, bts []byte) error { je.writeBin(bts); return nil },
		Bool: func(ctx *LensContext, bts []byte, data bool) error {
			if data {
				je.buf.WriteString("true")
			} else {
				je.buf.WriteString("false")
			}
			return nil
		},

		EnterArray:  func(ctx *LensContext, prefix byte, cnt int) error { je.buf.WriteByte('['); return nil },
		LeaveArray:  func(ctx *LensContext, prefix byte, cnt int, bts []byte) error { je.buf.WriteByte(']'); return nil },
		EnterMap:    func(ctx *LensContext, prefix byte, cnt int) error { je.buf.WriteByte('{'); return nil },
		LeaveMapKey: func(ctx *LensContext, n, cnt int) error { je.buf.WriteByte(':'); return nil },
		LeaveMap:    func(ctx *LensContext, prefix byte, cnt int, bts []byte) error { je.buf.WriteByte('}'); return nil },

		LeaveArrayElem: func(ctx *LensContext, n, cnt int) error {
			if n < cnt-1 {
				je.buf.WriteByte(',')
			}
			return nil
		},

		LeaveMapElem: func(ctx *LensContext, n, cnt int) error {
			if n < cnt-1 {
				je.buf.WriteByte(',')
			}
			return nil
		},
	}
	return je
}

func (j *JSONEncoder) writeBin(data []byte) {
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	str := *(*string)(unsafe.Pointer(&reflect.StringHeader{Data: sh.Data, Len: sh.Len}))
	j.writeJSONString(str)
}
