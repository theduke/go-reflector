package reflector

import (
	"errors"
	"reflect"
)

type StructReflector interface {
	// Interface returns the struct as interface{}.
	Interface() interface{}

	// Value returns the raw reflect.Value.
	Value() Reflector

	// Type returns the raw reflect.Type.
	Type() reflect.Type

	// New creates a new instance of the struct and returns an reflector.
	New() StructReflector

	// Returns a Reflector for a field, or nil if the field does not exist.
	Field(fieldName string) Reflector

	// Fields returns a map of fields, allowing you to easily iterate over all fields.
	Fields() map[string]Reflector

	// HasField returns true if the struct has a field with the specified name.
	HasField(fieldName string) bool

	// FieldValue returns the value of a struct field, or an error if the field
	// does not exist.
	FieldValue(fieldName string) (interface{}, error)

	// UFieldValue returns the value of a struct field, or nil if the field
	// does not exist.
	UFieldValue(fieldName string) interface{}

	SetFieldValue(fieldName string, value interface{}, convert ...bool) error
	SetField(fieldName string, value Reflector, convert ...bool) error

	// ToMap recursively converts the struct to a map[string]interface{} map.
	// You can optionally omit zero or empty values.
	ToMap(omitZero, omitEmpty bool) map[string]interface{}

	// FromMap sets struct fields from a map[string]interface{} map.
	//
	// You can optionally enable conversion of types that do not match by
	// passing true as a second argument.
	//
	// An error will be returned if values have a type mismatch, or, if
	// conversion is enabled, if a field conversion fails.
	FromMap(data map[string]interface{}, convert ...bool) error
}

type structReflector struct {
	item       Reflector
	structItem Reflector
	isPtr      bool
}

// Ensure that structReflector implements StructReflector.
var _ StructReflector = (*structReflector)(nil)

// Struct builds a new StructReflector.
// You may pass in a struct or a pointer to a struct.
func newStructReflector(v Reflector) (StructReflector, error) {
	// Dereference interfaces.
	if v.IsInterface() {
		if v.IsNil() {
			return nil, errors.New(ERR_INVALID_VALUE)
		}
		v = v.Elem()
	}

	if v.IsStruct() {
		return &structReflector{
			item:       v,
			structItem: v,
		}, nil
	} else if v.IsStructPtr() {
		if v.IsNil() {
			newStruct := New(v.Type().Elem())
			if err := v.Set(newStruct); err != nil {
				return nil, err
			}
		}

		return &structReflector{
			item:       v,
			structItem: v.Elem(),
			isPtr:      true,
		}, nil
	}
	return nil, errors.New(ERR_NOT_A_STRUCT)
}

func (r *structReflector) Interface() interface{} {
	return r.structItem.Interface()
}

func (r *structReflector) Value() Reflector {
	return r.structItem
}

func (r *structReflector) Type() reflect.Type {
	return r.structItem.Type()
}

func (r *structReflector) New() StructReflector {
	ptr := New(r.structItem.Type())
	refl, err := newStructReflector(ptr)

	// This should never happen, excpet when out of memory!
	// Panic is there just to make sure.
	if err != nil {
		panic(err)
	}

	return refl
}

func (r *structReflector) Field(fieldName string) Reflector {
	if !r.HasField(fieldName) {
		return nil
	}
	field := r.structItem.Value().FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	return Reflect(field)
}

func (r *structReflector) Fields() map[string]Reflector {
	m := make(map[string]Reflector)
	for i := 0; i < r.Type().NumField(); i++ {
		f := r.Type().Field(i)
		m[f.Name] = Reflect(r.structItem.Value().Field(i))
	}
	return m
}

func (r *structReflector) HasField(fieldName string) bool {
	_, ok := r.Type().FieldByName(fieldName)
	return ok
}

func (r *structReflector) FieldValue(fieldName string) (interface{}, error) {
	if !r.HasField(fieldName) {
		return nil, errors.New(ERR_UNKNOWN_FIELD)
	}

	field := r.item.Value().FieldByName(fieldName)
	if !field.IsValid() {
		return nil, errors.New(ERR_INVALID_FIELD)
	}
	if !field.CanInterface() {
		return nil, errors.New(ERR_UNINTERFACEABLE_FIELD)
	}
	return field.Interface(), nil
}

func (r *structReflector) UFieldValue(fieldName string) interface{} {
	v, err := r.FieldValue(fieldName)
	if err != nil {
		return nil
	}
	return v
}

func (r *structReflector) SetFieldValue(fieldName string, value interface{}, convert ...bool) error {
	v := Reflect(value)
	if v == nil {
		return errors.New(ERR_INVALID_VALUE)
	}
	return r.SetField(fieldName, v, convert...)
}

func (r *structReflector) SetField(fieldName string, value Reflector, convert ...bool) error {
	field := r.Field(fieldName)
	if field == nil {
		return errors.New(ERR_UNKNOWN_FIELD)
	}
	return field.Set(value, convert...)
}

func (r *structReflector) ToMap(omitZero, omitEmpty bool) map[string]interface{} {
	data := make(map[string]interface{})
	for name, field := range r.Fields() {
		if (field.IsStruct() || field.IsStructPtr()) && !field.IsZero() {
			s, _ := newStructReflector(field)
			d := s.ToMap(omitZero, omitEmpty)

			// Add embedded fields to the main data.
			if f, _ := r.Type().FieldByName(name); f.Anonymous {
				for key, val := range d {
					data[key] = val
				}
			} else {
				data[name] = d
			}
			continue
		}

		if omitEmpty && field.IsEmpty() {
			continue
		}
		if field.IsZero() {
			if omitZero {
				continue
			} else {
				data[name] = nil
				continue
			}
		}
		data[name] = field.Interface()
	}
	return data
}

func (r *structReflector) FromMap(data map[string]interface{}, convert ...bool) error {
	for key, rawVal := range data {
		field := r.Field(key)
		if field == nil {
			continue
		}

		val := Reflect(rawVal)
		if val == nil || val.IsZero() {
			continue
		}

		// Handle nested structs.
		if nestedMap, ok := rawVal.(map[string]interface{}); ok && field.IsStruct() || field.IsStructPtr() {
			// Obtain StructReflector.
			nestedStruct, err := field.Struct()
			if err != nil {
				return err
			}
			// run FromMap on nested struct.
			if err := nestedStruct.FromMap(nestedMap, convert...); err != nil {
				return err
			}

			// nested fromMap succeeded
			continue
		}

		// Handle regular values.
		if err := field.Set(val, convert...); err != nil {
			return errors.New("Error in field " + key + ": " + err.Error())
		}
	}
	return nil
}
