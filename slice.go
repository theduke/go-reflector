package reflector

import (
	"errors"
	"reflect"
)

type SliceReflector interface {
	String() string

	// Interface returns the underlying value as interface{}.
	//
	// IMPORTANT: If you created the slice from a pointer, or a pointer to a pointer,
	// this will not be a slice, but the pointer!
	// To obtain the slice, use SliceInterface()!
	Interface() interface{}

	// SliceInterface returns the slice as interface{}.
	SliceInterface() interface{}

	Value() Reflector

	// ItemType returns the type of slice items.
	ItemType() reflect.Type

	// Len returns the current length of the slice.
	Len() int

	// Index returns the item at the given index as a Reflector, or nil if the
	// index does not exist.
	Index(i int) Reflector
}

type sliceReflector struct {
	value      Reflector
	sliceValue reflect.Value
	canAppend  bool
}

// Ensure sliceReflector implements SliceReflector.
var _ SliceReflector = (*sliceReflector)(nil)

func slice(value Reflector) (SliceReflector, error) {
	if value.IsSlice() {
		// Direct slice.
		return &sliceReflector{
			value:      value,
			sliceValue: value.Value(),
		}, nil
	}

	if value.IsPtr() && value.Type().Elem().Kind() == reflect.Slice {
		if value.IsNil() {
			return nil, errors.New("nil_slice_ptr: Can't get a slice reflector for a nil slice pointer. Must pass pointer to pointer to allow auto-initialization!")
		} else {
			return &sliceReflector{
				value:      value,
				sliceValue: value.Elem().Value(),
			}, nil
		}
	}

	if value.IsPtr() && value.Type().Elem().Kind() == reflect.Ptr && value.Type().Elem().Elem().Kind() == reflect.Slice {
		// Got a pointer to a pointer to a slice.

		// Check if it is nil.
		if value.IsNil() {
			return nil, errors.New("nil_slice_ptr: Can't get a SliceReflector for a nil pointer to a pointer to a slice: must pass initialized pointer")
		}

		ptr := value.Elem()
		if ptr.IsNil() {
			// Ptr is nil, so create a new slice.

			// First Elem() gets the slice, second one the type of slice items.
			sliceItemType := ptr.Type().Elem().Elem()

			newSlicePtr := New(reflect.SliceOf(sliceItemType))

			if err := ptr.Set(newSlicePtr); err != nil {
				// Could not set pointer to new slice.
				return nil, err
			}

			return &sliceReflector{
				value:      value,
				sliceValue: newSlicePtr.Elem().Value(),
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
	return s.value.Interface()
}

func (s *sliceReflector) SliceInterface() interface{} {
	return s.sliceValue.Interface()
}

func (s *sliceReflector) Value() Reflector {
	return s.value
}

func (s *sliceReflector) ItemType() reflect.Type {
	return s.value.Elem().Type()
}

func (s *sliceReflector) Len() int {
	return s.sliceValue.Len()
}

func (s *sliceReflector) Index(i int) Reflector {
	if s.Len()-1 < i {
		return nil
	}
	return Reflect(s.value.Value().Index(i))
}
