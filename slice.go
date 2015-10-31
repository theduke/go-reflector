package reflector

import (
	"errors"
	"reflect"
)

type SliceReflector interface {
	String() string

	// Interface returns the slice as interface{}.
	Interface() interface{}

	Value() Reflector

	// Type returns the type of slice items.
	Type() reflect.Type

	// Len returns the current length of the slice.
	Len() int

	// Index returns the item at the given index as a Reflector, or nil if the
	// index does not exist.
	Index(index int) Reflector

	// Index returns the item at the given index as interface{}, or nil if the
	// index does not exist.
	IndexValue(index int) interface{}

	// Items returns a slice of Reflectors for each item in the array.
	// Allows easy iteration over all items.
	Items() []Reflector

	// Append appends an item to the slice.
	// Can only be used if the SliceReflector was created from a pointer to a slice.
	//
	// Returns an error if the value is of a different type than the slice, or if
	// the SliceReflector was not created from a pointer.
	Append(value ...Reflector) error

	// AppendValue appends an item to the slice.
	// Can only be used if the SliceReflector was created from a pointer to a slice.
	//
	// Returns an error if the value is of a different type than the slice, or if
	// the SliceReflector was not created from a pointer.
	AppendValue(value ...interface{}) error

	// ConvertTo converts the slice to a slice of a different type.
	// This requires that the slice items can be converted with normal Go conversion.
	//
	// This can, for example, be used to turn a slice like []interface{} into []string,
	// if the first slice only contains strings.
	ConvertTo(value interface{}) (slice interface{}, err error)

	// ConvertTo converts the slice to a slice of a different type.
	// This requires that the slice items can be converted with normal Go conversion.
	//
	// This can, for example, be used to turn a slice like []interface{} into []string,
	// if the first slice only contains strings.
	ConvertToType(typ reflect.Type) (slice interface{}, err error)
}

type sliceReflector struct {
	value      Reflector
	sliceValue Reflector
	canAppend  bool
}

// Ensure sliceReflector implements SliceReflector.
var _ SliceReflector = (*sliceReflector)(nil)

func newSliceReflector(value Reflector) (SliceReflector, error) {
	if value.IsSlice() {
		// Direct slice.
		return &sliceReflector{
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

			return &sliceReflector{
				value:      value,
				sliceValue: value.Elem(),
				canAppend:  true,
			}, nil
		}
	}

	return nil, errors.New(ERR_NOT_A_SLICE)
}

func (s *sliceReflector) String() string {
	return s.value.String()
}

func (s *sliceReflector) Interface() interface{} {
	return s.sliceValue.Interface()
}

func (s *sliceReflector) Value() Reflector {
	return s.value
}

func (s *sliceReflector) Type() reflect.Type {
	return s.sliceValue.Type().Elem()
}

func (s *sliceReflector) Len() int {
	return s.sliceValue.Len()
}

func (s *sliceReflector) Index(i int) Reflector {
	if i > s.Len()-1 {
		return nil
	}
	return Reflect(s.sliceValue.Value().Index(i))
}

func (s *sliceReflector) IndexValue(i int) interface{} {
	if v := s.Index(i); v != nil {
		return v.Interface()
	} else {
		return nil
	}
}

func (s *sliceReflector) Items() []Reflector {
	sl := make([]Reflector, 0)
	for i := 0; i < s.Len(); i++ {
		item := s.Index(i)
		if item.IsInterface() && !item.IsNil() {
			item = item.Elem()
		}
		sl = append(sl, item)
	}
	return sl
}

func (s *sliceReflector) Append(values ...Reflector) error {
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

func (s *sliceReflector) AppendValue(values ...interface{}) error {
	for _, val := range values {
		if err := s.Append(Reflect(val)); err != nil {
			return err
		}
	}
	return nil
}

func (s *sliceReflector) ConvertTo(value interface{}) (interface{}, error) {
	r := Reflect(value)
	if r == nil {
		return nil, errors.New(ERR_INVALID_VALUE)
	}
	return s.ConvertToType(r.Type())
}

func (s *sliceReflector) ConvertToType(typ reflect.Type) (interface{}, error) {
	newSlice := New(typ).Elem().NewSlice()
	if s.Len() == 0 {
		return newSlice, nil
	}

	for _, item := range s.Items() {
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
