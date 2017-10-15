package msgplens

import "fmt"

// Visitor provides functions that will be called as a msgpack object
// is walked for each different kind of child object that is encountered.
type Visitor struct {
	Begin func(ctx *LensContext) error
	End   func(ctx *LensContext, left []byte) error

	Str     func(ctx *LensContext, bts []byte, str string) error
	Int     func(ctx *LensContext, bts []byte, i int64) error
	Uint    func(ctx *LensContext, bts []byte, u uint64) error
	Bin     func(ctx *LensContext, bts []byte, bin []byte) error
	Float32 func(ctx *LensContext, bts []byte, f float32) error
	Float64 func(ctx *LensContext, bts []byte, f float64) error
	Bool    func(ctx *LensContext, bts []byte, b bool) error
	Nil     func(ctx *LensContext, prefix byte) error

	EnterArray     func(ctx *LensContext, prefix byte, cnt int) error
	EnterArrayElem func(ctx *LensContext, n, cnt int) error
	LeaveArrayElem func(ctx *LensContext, n, cnt int) error
	LeaveArray     func(ctx *LensContext, prefix byte, cnt int, bts []byte) error

	EnterMap     func(ctx *LensContext, prefix byte, cnt int) error
	EnterMapKey  func(ctx *LensContext, n, cnt int) error
	LeaveMapKey  func(ctx *LensContext, n, cnt int) error
	EnterMapElem func(ctx *LensContext, n, cnt int) error
	LeaveMapElem func(ctx *LensContext, n, cnt int) error
	LeaveMap     func(ctx *LensContext, prefix byte, cnt int, bts []byte) error

	Extension func(ctx *LensContext, bts []byte) error
}

type Visitable interface {
	Visitor() *Visitor
}

// WalkBytes walks the bytes in a msgpack object and visits each of the types
// using the Visitor created by a Visitable.
func WalkBytes(v Visitable, bts []byte) error {
	ctx := &LensContext{
		bts: bts,
		cur: 0,
		cnt: len(bts),
		vis: v.Visitor(),
	}
	return ctx.walkRoot()
}

type LensContext struct {
	bts  []byte
	last int
	cur  int
	cnt  int

	vis *Visitor
}

func (c *LensContext) Len() int { return c.cnt }
func (c *LensContext) Pos() int { return c.last }

func (c *LensContext) walkRoot() error {
	if c.vis.Begin != nil {
		if err := c.vis.Begin(c); err != nil {
			return err
		}
	}
	if err := c.walk(); err != nil {
		return fmt.Errorf("walk failed at position %d/%d: %v", c.cur, c.cnt, err)
	}
	if c.vis.End != nil {
		if err := c.vis.End(c, c.bts[c.cur:]); err != nil {
			return err
		}
	}
	return nil
}

func (c *LensContext) walkArray(contents []byte, objs uintptr) error {
	start := c.last
	if c.vis.EnterArray != nil {
		if err := c.vis.EnterArray(c, contents[0], int(objs)); err != nil {
			return err
		}
	}
	for i := 0; i < int(objs); i++ {
		if c.vis.EnterArrayElem != nil {
			if err := c.vis.EnterArrayElem(c, i, int(objs)); err != nil {
				return nil
			}
		}
		if err := c.walk(); err != nil {
			return err
		}
		if c.vis.LeaveArrayElem != nil {
			if err := c.vis.LeaveArrayElem(c, i, int(objs)); err != nil {
				return nil
			}
		}
	}
	if c.vis.LeaveArray != nil {
		if err := c.vis.LeaveArray(c, contents[0], int(objs), c.bts[start:c.cur]); err != nil {
			return err
		}
	}
	return nil
}

func (c *LensContext) walkMap(contents []byte, objs uintptr) error {
	start := c.last
	lim := int(objs) / 2

	if c.vis.EnterMap != nil {
		if err := c.vis.EnterMap(c, contents[0], lim); err != nil {
			return err
		}
	}
	for i := 0; i < lim; i++ {
		if c.vis.EnterMapKey != nil {
			if err := c.vis.EnterMapKey(c, i, lim); err != nil {
				return nil
			}
		}
		if err := c.walk(); err != nil {
			return err
		}
		if c.vis.LeaveMapKey != nil {
			if err := c.vis.LeaveMapKey(c, i, lim); err != nil {
				return nil
			}
		}
		if c.vis.EnterMapElem != nil {
			if err := c.vis.EnterMapElem(c, i, lim); err != nil {
				return nil
			}
		}
		if err := c.walk(); err != nil {
			return err
		}
		if c.vis.LeaveMapElem != nil {
			if err := c.vis.LeaveMapElem(c, i, lim); err != nil {
				return nil
			}
		}
	}
	if c.vis.LeaveMap != nil {
		if err := c.vis.LeaveMap(c, contents[0], lim, c.bts[start:c.cur]); err != nil {
			return err
		}
	}
	return nil
}

func (c *LensContext) walk() error {
	if c.cur >= c.cnt {
		return fmt.Errorf("tried to walk past end")
	}

	typ := getType(c.bts[c.cur])
	prefix := c.bts[c.cur]
	sz, objs, err := getSize(c.bts[c.cur:])
	if err != nil {
		return err
	}

	contents := c.bts[c.cur : c.cur+int(sz)]
	c.last = c.cur
	c.cur += int(sz)

	switch typ {
	case ArrayType:
		if err := c.walkArray(contents, objs); err != nil {
			return err
		}
	case MapType:
		if err := c.walkMap(contents, objs); err != nil {
			return err
		}
	case StrType:
		if c.vis.Str != nil {
			idx := 1
			idx += -int(sizes[prefix].extra)
			if err := c.vis.Str(c, contents, string(contents[idx:])); err != nil {
				return err
			}
		}
	case IntType:
		if c.vis.Int != nil {
			i, err := readInt64(contents)
			if err != nil {
				return err
			}
			if err := c.vis.Int(c, contents, i); err != nil {
				return err
			}
		}
	case UintType:
		if c.vis.Int != nil {
			u, err := readUint64(contents)
			if err != nil {
				return err
			}
			if err := c.vis.Uint(c, contents, u); err != nil {
				return err
			}
		}
	case BinType:
		if c.vis.Bin != nil {
			idx := 1
			idx += -int(sizes[prefix].extra)
			if err := c.vis.Bin(c, contents, contents[idx:]); err != nil {
				return err
			}
		}
	case BoolType:
		if c.vis.Bool != nil {
			if err := c.vis.Bool(c, contents, contents[0] == True); err != nil {
				return err
			}
		}
	case NilType:
		if c.vis.Nil != nil {
			if err := c.vis.Nil(c, Nil); err != nil {
				return err
			}
		}
	case Float64Type:
		if c.vis.Float64 != nil {
			f, err := readFloat64(contents)
			if err != nil {
				return err
			}
			if err := c.vis.Float64(c, contents, f); err != nil {
				return err
			}
		}
	case Float32Type:
		if c.vis.Float32 != nil {
			f, err := readFloat32(contents)
			if err != nil {
				return err
			}
			if err := c.vis.Float32(c, contents, f); err != nil {
				return err
			}
		}
	case ExtensionType:
		if c.vis.Extension != nil {
			if err := c.vis.Extension(c, contents); err != nil {
				return err
			}
		}
	}

	return nil
}
