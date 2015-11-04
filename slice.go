package reflector

import (
	"errors"
	"reflect"
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
	sl := make([]*Reflector, 0)
	for i := 0; i < s.Len(); i++ {
		item := s.Index(i)
		if item.IsInterface() && !item.IsNil() {
			item = item.Elem()
		}
		sl = append(sl, item)
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
func (s *SliceReflector) FilterBy(filterFunc func(sliceItem *Reflector) (include bool)) *SliceReflector {
	if s.Len() < 1 {
		return s
	}

	newSlice := s.Index(0).NewSlice()

	for _, item := range s.Items() {
		if filterFunc(item) {
			newSlice.Append(item)
		}
	}

	return newSlice
}
