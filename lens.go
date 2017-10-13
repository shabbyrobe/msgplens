package msgplens

type Visitor struct {
	Str   func(bts []byte, data string) error
	Int   func(bts []byte, data int64) error
	Uint  func(bts []byte, data uint64) error
	Bin   func(bts []byte, data []byte) error
	Float func(bts []byte, data float64) error
	Bool  func(bts []byte, data bool) error
	Nil   func() error

	EnterArray     func(prefix byte, len int) error
	EnterArrayElem func(n, cnt int) error
	LeaveArrayElem func(n, cnt int) error
	LeaveArray     func() error

	EnterMap     func(prefix byte, len int) error
	EnterMapKey  func(n, cnt int) error
	LeaveMapKey  func(n, cnt int) error
	EnterMapElem func(n, cnt int) error
	LeaveMapElem func(n, cnt int) error
	LeaveMap     func() error

	Extension func(bts []byte) error
}

func WalkBytes(v *Visitor, bts []byte) error {
	ctx := &ctx{
		bts:   bts,
		cur:   0,
		cnt:   len(bts),
		vis:   v,
		stack: make([]walkStack, 0, 8),
		level: -1,
	}
	return ctx.walk()
}

type ctx struct {
	bts  []byte
	left []byte
	cur  int
	cnt  int
	vis  *Visitor

	stack []walkStack
	level int
}

func (c *ctx) push(typ Type, objs uintptr) {
	c.stack = append(c.stack, walkStack{typ, objs, objs})
	c.level++
}

func (c *ctx) pop() (out walkStack) {
	c.stack, out = c.stack[0:c.level], c.stack[c.level]
	c.level--
	return
}

func (c *ctx) leave() error {
	for c.level >= 0 && c.stack[c.level].n == 0 {
		w := c.pop()
		switch w.typ {
		case ArrayType:
			if c.vis.LeaveArray != nil {
				if err := c.vis.LeaveArray(); err != nil {
					return err
				}
			}
		case MapType:
			if c.vis.LeaveMap != nil {
				if err := c.vis.LeaveMap(); err != nil {
					return err
				}
			}
		}
		if c.level >= 0 {
			c.stack[c.level].n--
		}
	}
	return nil
}

func (c *ctx) enterChild() error {
	cnt := int(c.stack[c.level].cnt)
	n := cnt - int(c.stack[c.level].n)

	switch c.stack[c.level].typ {
	case ArrayType:
		if c.vis.EnterArrayElem != nil {
			if err := c.vis.EnterArrayElem(n, cnt); err != nil {
				return err
			}
		}

	case MapType:
		cnt /= 2
		n /= 2
		if c.stack[c.level].n%2 == 0 {
			if c.vis.EnterMapKey != nil {
				if err := c.vis.EnterMapKey(n, cnt); err != nil {
					return err
				}
			}
		} else {
			if c.vis.EnterMapElem != nil {
				if err := c.vis.EnterMapElem(n, cnt); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *ctx) leaveChild() error {
	cnt := int(c.stack[c.level].cnt)
	n := cnt - int(c.stack[c.level].n)

	switch c.stack[c.level].typ {
	case ArrayType:
		if c.vis.LeaveArrayElem != nil {
			if err := c.vis.LeaveArrayElem(n, cnt); err != nil {
				return err
			}
		}

	case MapType:
		cnt /= 2
		n /= 2
		if c.stack[c.level].n%2 == 0 {
			if c.vis.LeaveMapKey != nil {
				if err := c.vis.LeaveMapKey(n, cnt); err != nil {
					return err
				}
			}
		} else {
			if c.vis.LeaveMapElem != nil {
				if err := c.vis.LeaveMapElem(n, cnt); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *ctx) walk() error {
	c.left = c.bts

	for c.cur < c.cnt {
		typ := getType(c.left[0])
		bts, objs, err := getSize(c.left)
		if err != nil {
			return err
		}

		c.cur += int(bts)
		contents := c.left[0:bts]
		c.left = c.left[bts:]
		pushed := false

		if c.level >= 0 {
			if err := c.enterChild(); err != nil {
				return err
			}
		}

		switch typ {
		case ArrayType:
			c.push(ArrayType, objs)
			pushed = true
			if c.vis.EnterArray != nil {
				if err := c.vis.EnterArray(contents[0], int(objs)); err != nil {
					return err
				}
			}
		case MapType:
			c.push(MapType, objs)
			pushed = true
			if c.vis.EnterMap != nil {
				if err := c.vis.EnterMap(contents[0], int(objs)); err != nil {
					return err
				}
			}
		case StrType:
			if c.vis.Str != nil {
				if err := c.vis.Str(contents, string(contents[1:])); err != nil {
					return err
				}
			}
		case IntType:
			if c.vis.Int != nil {
				i, err := readInt64(contents)
				if err != nil {
					return err
				}
				if err := c.vis.Int(contents, i); err != nil {
					return err
				}
			}
		case UintType:
			if c.vis.Int != nil {
				u, err := readUint64(contents)
				if err != nil {
					return err
				}
				if err := c.vis.Uint(contents, u); err != nil {
					return err
				}
			}
		case BinType:
			if c.vis.Bin != nil {
				if err := c.vis.Bin(contents, contents[1:]); err != nil {
					return err
				}
			}
		case BoolType:
			if c.vis.Bool != nil {
				if err := c.vis.Bool(contents, contents[0] == True); err != nil {
					return err
				}
			}
		case NilType:
			if c.vis.Nil != nil {
				if err := c.vis.Nil(); err != nil {
					return err
				}
			}
		case Float64Type:
			if c.vis.Float != nil {
				f, err := readFloat64(contents)
				if err != nil {
					return err
				}
				if err := c.vis.Float(contents, f); err != nil {
					return err
				}
			}
		case Float32Type:
			if c.vis.Float != nil {
				f, err := readFloat32(contents)
				if err != nil {
					return err
				}
				if err := c.vis.Float(contents, float64(f)); err != nil {
					return err
				}
			}
		case ExtensionType:
			if c.vis.Extension != nil {
				if err := c.vis.Extension(contents); err != nil {
					return err
				}
			}
		}
		if !pushed {
			if c.level >= 0 {
				if err := c.leaveChild(); err != nil {
					return err
				}
			}
			c.stack[c.level].n--
		}
		c.leave()
	}

	c.leave()

	return nil
}

type walkStack struct {
	typ Type
	n   uintptr
	cnt uintptr
}
