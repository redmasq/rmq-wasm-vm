package wasmvm

type ValueStackEntryType int8

const (
	TYPE_I32 ValueStackEntryType = iota
	TYPE_F32
	TYPE_I64
	TYPE_F64
)

type ValueStackEntry struct {
	EntryType ValueStackEntryType
	Value_I32 uint32
	Value_F32 float32
	Value_I64 uint64
	Value_F64 float64
}

type ValueStack struct {
	elements []ValueStackEntry
}

func NewValueStack() *ValueStack {
	return &ValueStack{
		elements: make([]ValueStackEntry, 0),
	}
}

func NewValueStackEntryI32(value uint32) *ValueStackEntry {
	return &ValueStackEntry{
		EntryType: TYPE_I32,
		Value_I32: value,
	}
}

func NewValueStackEntryI64(value uint64) *ValueStackEntry {
	return &ValueStackEntry{
		EntryType: TYPE_I64,
		Value_I64: value,
	}
}

func (vs *ValueStack) Push(item *ValueStackEntry) {
	vs.elements = append(vs.elements, *item)
}

func (vs *ValueStack) PushInt32(item uint32) {
	stackEntry := NewValueStackEntryI32(item)
	vs.Push(stackEntry)
}

func (vs *ValueStack) PushInt64(item uint64) {
	stackEntry := NewValueStackEntryI64(item)
	vs.Push(stackEntry)
}

func (vs *ValueStack) IsEmpty() bool {
	return len(vs.elements) == 0
}

func (vs *ValueStack) HasAtLeast(cnt int) bool {
	return len(vs.elements) >= cnt
}

func (vs *ValueStack) Size() int {
	return len(vs.elements)
}

func (vs *ValueStack) HasAtLeastOfType(cnt int, entryType ValueStackEntryType) (bool, []ValueStackEntry) {
	if !vs.HasAtLeast(cnt) {
		return false, nil
	}
	n := len(vs.elements) - cnt
	items := vs.elements[n:]
	for _, val := range items {
		if val.EntryType != entryType {
			return false, nil
		}
	}
	return true, items
}

func (vs *ValueStack) Drop(cnt int, allOrNothing bool) bool {
	if (allOrNothing && !vs.HasAtLeast(cnt)) || vs.IsEmpty() {
		return false
	}
	n := len(vs.elements) - cnt
	vs.elements = vs.elements[:n] // Slice off the last element
	return true
}

func (vs *ValueStack) Pop() (*ValueStackEntry, bool) {
	if vs.IsEmpty() {
		return nil, false
	}
	n := len(vs.elements) - 1
	item := vs.elements[n]
	vs.elements = vs.elements[:n] // Slice off the last element
	return &item, true
}
