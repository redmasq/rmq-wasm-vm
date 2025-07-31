package wasmvm_test

import (
	"strconv"
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValueStack(t *testing.T) {
	vs := wasmvm.NewValueStack()
	require.NotNil(t, vs)
	assert.Equal(t, 0, vs.Size())
}

func TestValueStack_PushAndPop(t *testing.T) {
	type testcase struct {
		name       string
		push       func(vs *wasmvm.ValueStack)
		expectedET wasmvm.ValueStackEntryType
		expectedV  interface{}
	}

	cases := []testcase{
		{
			name: "Push and Pop I32",
			push: func(vs *wasmvm.ValueStack) {
				vs.PushInt32(0xDEADBEEF)
			},
			expectedET: wasmvm.TYPE_I32,
			expectedV:  uint32(0xDEADBEEF),
		},
		{
			name: "Push and Pop I64",
			push: func(vs *wasmvm.ValueStack) {
				vs.PushInt64(0xDEADBEEFCAFED00D)
			},
			expectedET: wasmvm.TYPE_I64,
			expectedV:  uint64(0xDEADBEEFCAFED00D),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vs := wasmvm.NewValueStack()
			c.push(vs)
			assert.Equal(t, 1, vs.Size())

			item, ok := vs.Pop()
			require.True(t, ok)
			require.NotNil(t, item)
			assert.Equal(t, c.expectedET, item.EntryType)
			switch c.expectedET {
			case wasmvm.TYPE_I32:
				assert.Equal(t, c.expectedV, item.Value_I32)
			case wasmvm.TYPE_I64:
				assert.Equal(t, c.expectedV, item.Value_I64)
			}
			assert.True(t, vs.IsEmpty())
		})
	}
}

func TestValueStack_Drop(t *testing.T) {
	cases := []struct {
		name     string
		pushVals []int32
		dropN    int
		dropOK   bool
		expected int
	}{
		{
			name:     "Drop one from two",
			pushVals: []int32{42, 99},
			dropN:    1,
			dropOK:   true,
			expected: 1,
		},
		{
			name:     "Drop two from one",
			pushVals: []int32{42},
			dropN:    2,
			dropOK:   false,
			expected: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vs := wasmvm.NewValueStack()
			for _, v := range c.pushVals {
				vs.PushInt32(uint32(v))
			}
			ok := vs.Drop(c.dropN, true)
			assert.Equal(t, c.dropOK, ok)
			assert.Equal(t, c.expected, vs.Size())
		})
	}
}

func TestValueStack_HasAtLeast(t *testing.T) {
	cases := []struct {
		name     string
		actions  func(*wasmvm.ValueStack)
		n        int
		expected bool
	}{
		{
			name:     "HasAtLeast: Empty stack, needs one",
			actions:  func(vs *wasmvm.ValueStack) {},
			n:        1,
			expected: false,
		},
		{
			name:     "HasAtLeast: One i32, needs one",
			actions:  func(vs *wasmvm.ValueStack) { vs.PushInt32(123) },
			n:        1,
			expected: true,
		},
		{
			name:     "HasAtLeast: One i32, needs two",
			actions:  func(vs *wasmvm.ValueStack) { vs.PushInt32(123) },
			n:        2,
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vs := wasmvm.NewValueStack()
			c.actions(vs)
			assert.Equal(t, c.expected, vs.HasAtLeast(c.n))
		})
	}
}

func TestValueStack_HasAtLeastOfType(t *testing.T) {
	cases := []struct {
		name      string
		actions   func(*wasmvm.ValueStack)
		n         int
		expected  bool
		collect   []wasmvm.ValueStackEntry
		entryType wasmvm.ValueStackEntryType
	}{
		{
			name:      "HasAtLeastOfType: Empty stack, needs one",
			actions:   func(vs *wasmvm.ValueStack) {},
			n:         1,
			expected:  false,
			collect:   []wasmvm.ValueStackEntry{},
			entryType: wasmvm.TYPE_I32,
		},
		{
			name:     "HasAtLeastOfType: One i32, needs one",
			actions:  func(vs *wasmvm.ValueStack) { vs.PushInt32(123) },
			n:        1,
			expected: true,
			collect: []wasmvm.ValueStackEntry{
				{
					EntryType: wasmvm.TYPE_I32,
					Value_I32: uint32(123),
				},
			},
			entryType: wasmvm.TYPE_I32,
		},
		{
			name:      "HasAtLeastOfType: One i32, needs two",
			actions:   func(vs *wasmvm.ValueStack) { vs.PushInt32(123) },
			n:         2,
			expected:  false,
			collect:   nil,
			entryType: wasmvm.TYPE_I32,
		},
		{
			name:     "HasAtLeastOfType: One i64, needs one",
			actions:  func(vs *wasmvm.ValueStack) { vs.PushInt64(123) },
			n:        1,
			expected: true,
			collect: []wasmvm.ValueStackEntry{
				{
					EntryType: wasmvm.TYPE_I64,
					Value_I64: uint64(123),
				},
			},
			entryType: wasmvm.TYPE_I64,
		},
		{
			name: "HasAtLeastOfType: One i64 and two i32, needs two i32 (success)",
			actions: func(vs *wasmvm.ValueStack) {
				vs.PushInt64(0)
				vs.PushInt32(123)
				vs.PushInt32(456)
			},
			n:        2,
			expected: true,
			collect: []wasmvm.ValueStackEntry{
				{
					EntryType: wasmvm.TYPE_I32,
					Value_I32: uint32(123),
				},
				{
					EntryType: wasmvm.TYPE_I32,
					Value_I32: uint32(456),
				},
			},
			entryType: wasmvm.TYPE_I32,
		},
		{
			name: "HasAtLeastOfType: Two i64 and one i32, needs two i32 (failure)",
			actions: func(vs *wasmvm.ValueStack) {
				vs.PushInt64(0)
				vs.PushInt64(123)
				vs.PushInt32(456)
			},
			n:         2,
			expected:  false,
			collect:   nil,
			entryType: wasmvm.TYPE_I32,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vs := wasmvm.NewValueStack()
			c.actions(vs)
			success, collect := vs.HasAtLeastOfType(c.n, c.entryType)
			assert.Equal(t, c.expected, success)

			if c.collect == nil {
				assert.Nil(t, collect)
			} else if len(c.collect) == len(collect) {
				assert.Equal(t, len(c.collect), len(collect))
				for i := range c.collect {
					e := c.collect[i]
					x := collect[i]
					assert.Equal(t, e.EntryType, x.EntryType)
					switch e.EntryType {
					case wasmvm.TYPE_I32:
						assert.Equal(t, e.Value_I32, x.Value_I32)
					case wasmvm.TYPE_I64:
						assert.Equal(t, e.Value_I64, x.Value_I64)
					case wasmvm.TYPE_F32:
						assert.Equal(t, e.Value_F32, x.Value_F32)
					case wasmvm.TYPE_F64:
						assert.Equal(t, e.Value_F64, x.Value_F64)
					}
				}
			}

		})
	}
}

func TestValueStack_IsEmpty(t *testing.T) {
	vs := wasmvm.NewValueStack()
	assert.True(t, vs.IsEmpty())
	vs.PushInt32(7)
	assert.False(t, vs.IsEmpty())
}

func TestValueStack_Pop_Empty(t *testing.T) {
	vs := wasmvm.NewValueStack()
	val, ok := vs.Pop()
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestValueStack_Size(t *testing.T) {
	vs := wasmvm.NewValueStack()
	assert.Equal(t, 0, vs.Size())
	vs.PushInt32(1)
	assert.Equal(t, 1, vs.Size())
	vs.PushInt32(2)
	assert.Equal(t, 2, vs.Size())
	_, _ = vs.Pop()
	assert.Equal(t, 1, vs.Size())
}

func TestValueStackEntryString(t *testing.T) {
	cases := []struct {
		vType wasmvm.ValueStackEntryType
		name  string
	}{
		{
			vType: wasmvm.TYPE_I32,
			name:  "TYPE_I32",
		},
		{
			vType: wasmvm.TYPE_I64,
			name:  "TYPE_I64",
		},
		{
			vType: wasmvm.TYPE_F32,
			name:  "TYPE_F32",
		},
		{
			vType: wasmvm.TYPE_F64,
			name:  "TYPE_F64",
		},
		{
			vType: wasmvm.ValueStackEntryType(int(wasmvm.TYPE_F64) + 1),
			name:  "ValueStackEntryType(" + strconv.Itoa(int(wasmvm.TYPE_F64)+1) + ")",
		},
	}

	for i := range cases {

		c := cases[i]
		t.Run("Correct enum string: "+c.name, func(t *testing.T) {
			s := c.vType.String()
			assert.Equal(t, c.name, s)
		})
	}
}

func TestErrorGenerators(t *testing.T) {
	memoryContent := []byte{wasmvm.OP_NOP}
	memorySize := uint64(len(memoryContent))
	cfg := &wasmvm.VMConfig{
		Size: memorySize,
		Image: &wasmvm.ImageConfig{
			Type:  wasmvm.Array,
			Array: memoryContent,
			Size:  memorySize,
		},
	}

	cases := []struct {
		reason string
		method wasmvm.ErrorGenerator
	}{
		{
			reason: "Stack Underflow",
			method: wasmvm.NewStackUnderflowErrorAndSetTrapReason,
		},
		{
			reason: "Stack Cleanup Error",
			method: wasmvm.NewStackCleanupErrorAndSetTrapReason,
		},
	}

	for i := range cases {

		c := cases[i]
		t.Run("VM Error Generation: "+c.reason, func(t *testing.T) {
			vm, err := wasmvm.NewVM(cfg)
			assert.NoError(t, err)
			err = c.method(vm, "dummy")
			assert.Error(t, err)
			assert.True(t, vm.Trap)
			assert.Equal(t, "dummy: "+c.reason, vm.TrapReason)
		})
	}

}
