package msgplens

func ExampleVisitor_template() {
	visitor := &Visitor{
		Str: func(ctx *LensContext, bts []byte, str string) error {
			return nil
		},
		Int: func(ctx *LensContext, bts []byte, data int64) error {
			return nil
		},
		Uint: func(ctx *LensContext, bts []byte, data uint64) error {
			return nil
		},
		Bin: func(ctx *LensContext, bts []byte, data []byte) error {
			return nil
		},
		Float: func(ctx *LensContext, bts []byte, data float64) error {
			return nil
		},
		Bool: func(ctx *LensContext, bts []byte, data bool) error {
			return nil
		},
		Nil: func(ctx *LensContext, prefix byte) error {
			return nil
		},
		Extension: func(ctx *LensContext, bts []byte) error {
			return nil
		},
		EnterArray: func(ctx *LensContext, prefix byte, len int) error {
			return nil
		},
		EnterArrayElem: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		LeaveArrayElem: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		LeaveArray: func(ctx *LensContext) error {
			return nil
		},

		EnterMap: func(ctx *LensContext, prefix byte, len int) error {
			return nil
		},
		EnterMapKey: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		LeaveMapKey: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		EnterMapElem: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		LeaveMapElem: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		LeaveMap: func(ctx *LensContext) error {
			return nil
		},
	}
}
