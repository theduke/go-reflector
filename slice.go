package reflector

import (
	"errors"
	"reflect"
	"sort"
)

type SliceReflector struct {
	value      *Reflector
	sliceValue *Reflector
	canAppend  bool
}

func newSliceReflector(value *Reflector) (*SliceReflector, error) {
	if !value.IsValid() {
		return nil, errors.New(ERR_INVALID_VALUE)
	}

	if value.IsSlice() {
		// Direct slice.
		return &SliceReflector{
			value:      value,
			sliceValue: value,
		}, nil
	}

	if value.IsPtr() && value.Type().Elem().Kind() == reflect.Slice {
		if value.IsNil() {
			return nil, errors.New("nil_slice_ptr: Can't get a slice reflector for a nil slice pointer. Must pass an initialized pointer!")
		} else {
			if value.Elem().IsNil() {
				sliceItemType := value.Type().Elem().Elem()
				slice := New(reflect.SliceOf(sliceItemType)).Elem()

				if err := value.Elem().Set(slice); err != nil {
					// Could not set pointer to new slice.
					return nil, err
				}
			}

			return &SliceReflector{
				value:      value,
				sliceValue: value.Elem(),
				canAppend:  true,
			}, nil
		}
	}

	return nil, errors.New(ERR_NOT_A_SLICE)
}

func (s *SliceReflector) String() string {
	return s.value.String()
}

func (s *SliceReflector) Interface() interface{} {
	return s.sliceValue.Interface()
}

func (s *SliceReflector) Value() *Reflector {
	return s.value
}

func (s *SliceReflector) Type() reflect.Type {
	return s.sliceValue.Type().Elem()
}

func (s *SliceReflector) Len() int {
	return s.sliceValue.Len()
}

func (s *SliceReflector) Cap() int {
	return s.sliceValue.Value().Cap()
}

func (s *SliceReflector) New() *SliceReflector {
	return New(s.Type()).NewSlice()
}

func (s *SliceReflector) Index(i int) *Reflector {
	if i > s.Len()-1 {
		return nil
	}
	return Reflect(s.sliceValue.Value().Index(i))
}

func (s *SliceReflector) IndexValue(i int) interface{} {
	if v := s.Index(i); v != nil {
		return v.Interface()
	} else {
		return nil
	}
}

func (s *SliceReflector) SetIndex(index int, value *Reflector) error {
	if index > s.Cap()-1 {
		return errors.New(ERR_INDEX_OUT_OF_BOUNDS)
	}
	if err := s.Index(index).Set(value); err != nil {
		return err
	}
	return nil
}

func (s *SliceReflector) SetIndexValue(index int, value interface{}) error {
	return s.SetIndex(index, Reflect(value))
}

func (s *SliceReflector) Swap(index1, index2 int) error {
	if index1 > s.Cap()-1 || index2 > s.Cap()-1 {
		return errors.New(ERR_INDEX_OUT_OF_BOUNDS)
	}
	v1 := s.Index(index1).Interface()
	v2 := s.Index(index2).Interface()
	if err := s.SetIndex(index1, Reflect(v2)); err != nil {
		return err
	}
	if err := s.SetIndex(index2, Reflect(v1)); err != nil {
		return err
	}
	return nil
}

func (s *SliceReflector) Items() []*Reflector {
	sl := make([]*Reflector, s.Len(), s.Len())
	for i := 0; i < s.Len(); i++ {
		item := s.Index(i)
		if item.IsInterface() && !item.IsNil() {
			item = item.Elem()
		}
		sl[i] = item
	}
	return sl
}

func (s *SliceReflector) Append(values ...*Reflector) error {
	if !s.canAppend {
		return errors.New(ERR_CANT_APPEND_NOT_A_POINTER)
	}

	var newSlice reflect.Value

	for _, val := range values {
		if val.Type() != s.Type() {
			return errors.New(ERR_TYPE_MISMATCH)
		}

		newSlice = reflect.Append(s.sliceValue.Value(), val.Value())
	}

	s.value.Elem().Set(Reflect(newSlice))
	return nil
}

func (s *SliceReflector) AppendValue(values ...interface{}) error {
	for _, val := range values {
		if err := s.Append(Reflect(val)); err != nil {
			return err
		}
	}
	return nil
}

func (s *SliceReflector) ConvertTo(value interface{}) (interface{}, error) {
	r := Reflect(value)
	if r == nil {
		return nil, errors.New(ERR_INVALID_VALUE)
	}
	return s.ConvertToType(r.Type())
}

func (s *SliceReflector) ConvertToType(typ reflect.Type) (interface{}, error) {
	newSlice := New(typ).Elem().NewSlice()
	if s.Len() == 0 {
		return newSlice, nil
	}

	for _, item := range s.Items() {
		// De-reference interfaces.
		if item.IsInterface() {
			item = item.Elem()
		}

		// If target is no pointer, dereference pointers.
		if item.IsPtr() && typ.Kind() != reflect.Ptr {
			item = item.Elem()
		}

		if item.Type() != typ {
			if !item.Type().ConvertibleTo(typ) {
				return nil, errors.New(ERR_TYPE_MISMATCH)
			}
			item = Reflect(item.Value().Convert(typ))
		}

		if err := newSlice.Append(item); err != nil {
			return nil, err
		}
	}

	return newSlice.Interface(), nil
}

// FilterBy filters the slice with a function that is called for each slice item, and must return true or false.
// Returns a new slice that only contains the filtered items.
func (s *SliceReflector) FilterBy(filterFunc func(sliceItem *Reflector) (include bool, err error)) (*SliceReflector, error) {
	if s.Len() < 1 {
		return s, nil
	}

	newSlice := s.Index(0).NewSlice()

	for _, item := range s.Items() {
		if include, err := filterFunc(item); err != nil {
			return nil, err
		} else if include {
			newSlice.Append(item)
		}
	}

	return newSlice, nil
}

func (s *SliceReflector) SortBy(sorterFunc func(a, b *Reflector) (bool, error)) error {
	if s.Len() < 1 {
		return nil
	}

	sorter := sliceSorter{
		slice:    s,
		items:    s.Items(),
		sortFunc: sorterFunc,
	}

	sort.Sort(sorter)
	if sorter.err != nil {
		return sorter.err
	}

	return nil
}

func (s *SliceReflector) SortByFieldFunc(fieldName string, sorterFunc func(a, b *Reflector) (bool, error)) error {
	if s.Len() < 1 {
		return nil
	}

	firstItem := s.Index(0)

	if !(firstItem.IsStructPtr() || firstItem.IsStruct() || firstItem.IsMap()) {
		return errors.New("Can't sort by field when slice items are neither pointers to structs, structs or maps")
	}

	if firstItem.IsStructPtr() || firstItem.IsStruct() {
		// Check that struct field exists.
		if !firstItem.MustStruct().HasField(fieldName) {
			return errors.New(ERR_UNKNOWN_FIELD)
		}
	}

	items := make([]*Reflector, s.Len(), s.Len())
	for i, item := range s.Items() {
		if item.IsStructPtr() {
			item = item.Elem()
		}
		items[i] = item
	}

	fieldNameVal := reflect.ValueOf(fieldName)

	sorter := sliceSorter{
		slice: s,
		items: s.Items(),
		sortFunc: func(a, b *Reflector) (bool, error) {
			var aVal, bVal *Reflector

			if a.IsStructPtr() {
				a = a.Elem()
			}
			if b.IsStructPtr() {
				b = b.Elem()
			}

			if a.IsStruct() {
				aVal = a.MustStruct().Field(fieldName)
				bVal = b.MustStruct().Field(fieldName)
			} else {
				aVal = R(a.Value().MapIndex(fieldNameVal))
				bVal = R(b.Value().MapIndex(fieldNameVal))
			}

			// De-reference interfaces.
			if aVal.IsInterface() && aVal.IsValid() {
				aVal = aVal.Elem()
			}
			if bVal.IsInterface() && bVal.IsValid() {
				bVal = bVal.Elem()
			}

			return sorterFunc(aVal, bVal)
		},
	}

	sort.Sort(sorter)
	if sorter.err != nil {
		return sorter.err
	}

	return nil
}

func (s *SliceReflector) SortByField(fieldName string, ascending bool) error {
	operator := "<"
	if !ascending {
		operator = ">"
	}

	sorter := func(a, b *Reflector) (bool, error) {
		return a.CompareTo(b, operator)
	}

	return s.SortByFieldFunc(fieldName, sorter)
}

type sliceSorter struct {
	slice    *SliceReflector
	items    []*Reflector
	sortFunc func(a, b *Reflector) (bool, error)
	err      error
}

func (s sliceSorter) Len() int {
	return s.slice.Len()
}

func (s sliceSorter) Swap(i, j int) {
	s.slice.Swap(i, j)
}

func (s sliceSorter) Less(i, j int) bool {
	flag, err := s.sortFunc(s.items[i], s.items[j])
	if err != nil {
		s.err = err
		return true
	}

	return flag
}
