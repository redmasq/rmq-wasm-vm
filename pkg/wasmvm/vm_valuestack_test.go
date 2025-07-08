package wasmvm_test

import (
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

func TestValueStack_PushInt32AndPop(t *testing.T) {
	vs := wasmvm.NewValueStack()
	vs.PushInt32(0xDEADBEEF)
	assert.Equal(t, 1, vs.Size())

	item, ok := vs.Pop()
	require.True(t, ok)
	require.NotNil(t, item)
	assert.Equal(t, wasmvm.TYPE_I32, item.EntryType)
	assert.Equal(t, uint32(0xDEADBEEF), item.Value_I32)
	assert.True(t, vs.IsEmpty())
}

func TestValueStack_PushInt64AndPop(t *testing.T) {
	vs := wasmvm.NewValueStack()
	vs.PushInt64(0xDEADBEEFCAFED00D)
	assert.Equal(t, 1, vs.Size())

	item, ok := vs.Pop()
	require.True(t, ok)
	require.NotNil(t, item)
	assert.Equal(t, wasmvm.TYPE_I64, item.EntryType)
	assert.Equal(t, uint64(0xDEADBEEFCAFED00D), item.Value_I64)
	assert.True(t, vs.IsEmpty())
}

func TestValueStack_HasAtLeastOfType(t *testing.T) {
	vs := wasmvm.NewValueStack()
	vs.PushInt32(1)
	vs.PushInt32(2)

	ok, items := vs.HasAtLeastOfType(2, wasmvm.TYPE_I32)
	assert.True(t, ok)
	require.Len(t, items, 2)
	assert.Equal(t, wasmvm.TYPE_I32, items[0].EntryType)
	assert.Equal(t, wasmvm.TYPE_I32, items[1].EntryType)

	vs.Push(&wasmvm.ValueStackEntry{EntryType: wasmvm.TYPE_F32, Value_F32: 1.0})
	ok, _ = vs.HasAtLeastOfType(2, wasmvm.TYPE_I32)
	assert.False(t, ok)
}

func TestValueStack_Drop(t *testing.T) {
	vs := wasmvm.NewValueStack()
	vs.PushInt32(42)
	vs.PushInt32(99)

	ok := vs.Drop(1, true)
	assert.True(t, ok)
	assert.Equal(t, 1, vs.Size())

	ok = vs.Drop(2, true)
	assert.False(t, ok) // Can't drop more than what's left
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

func TestValueStack_HasAtLeast(t *testing.T) {
	vs := wasmvm.NewValueStack()
	assert.False(t, vs.HasAtLeast(1))
	vs.PushInt32(123)
	assert.True(t, vs.HasAtLeast(1))
	assert.False(t, vs.HasAtLeast(2))
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
