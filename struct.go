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

	SetField(fieldName string, value interface{}, convert ...bool) error

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
	rawValue      interface{}
	rawValueIsPtr bool
	item          reflect.Value
	typ           reflect.Type
}

// Ensure that structReflector implements StructReflector.
var _ StructReflector = (*structReflector)(nil)

// Struct builds a new StructReflector.
// You may pass in a struct or a pointer to a struct.
func Struct(s interface{}) (StructReflector, error) {
	// Check for nil.
	if s == nil {
		return nil, errors.New(ERR_POINTER_OR_STRUCT_EXPECTED)
	}

	// Check if it is a pointer, and if so, dereference it.
	v := reflect.ValueOf(s)
	if !v.IsValid() {
		return nil, errors.New(ERR_INVALID_FIELD)
	}

	r := &structReflector{
		rawValue: s,
	}

	if v.Type().Kind() == reflect.Ptr {
		r.rawValueIsPtr = true
		v = v.Elem()
	}

	// Check that value is actually a struct.
	if v.Type().Kind() != reflect.Struct {
		return nil, errors.New(ERR_STRUCT_EXPECTED)
	}

	// Valid struct.
	r.item = v
	r.typ = v.Type()
	return r, nil
}

// MustStruct builds a new StructReflector, and panics if building the
// StructReflector fails.
func MustStruct(s interface{}) StructReflector {
	r, err := Struct(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *structReflector) Interface() interface{} {
	return r.rawValue
}

func (r *structReflector) Value() Reflector {
	return ReflectVal(r.item)
}

func (r *structReflector) Type() reflect.Type {
	return r.typ
}

func (r *structReflector) New() StructReflector {
	ptr := reflect.New(r.typ).Interface()
	return MustStruct(ptr)
}

func (r *structReflector) Field(fieldName string) Reflector {
	if !r.HasField(fieldName) {
		return nil
	}
	field := r.item.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	return ReflectVal(field)
}

func (r *structReflector) Fields() map[string]Reflector {
	m := make(map[string]Reflector)
	for i := 0; i < r.typ.NumField(); i++ {
		f := r.typ.Field(i)
		m[f.Name] = ReflectVal(r.item.Field(i))
	}
	return m
}

func (r *structReflector) HasField(fieldName string) bool {
	_, ok := r.typ.FieldByName(fieldName)
	return ok
}

func (r *structReflector) FieldValue(fieldName string) (interface{}, error) {
	if !r.HasField(fieldName) {
		return nil, errors.New(ERR_UNKNOWN_FIELD)
	}

	field := r.item.FieldByName(fieldName)
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

func (r *structReflector) SetField(fieldName string, value interface{}, convert ...bool) error {
	field := r.Field(fieldName)
	if field == nil {
		return errors.New(ERR_UNKNOWN_FIELD)
	}
	return field.SetValue(value, convert...)
}

func (r *structReflector) ToMap(omitZero, omitEmpty bool) map[string]interface{} {
	data := make(map[string]interface{})
	for name, field := range r.Fields() {
		if (field.IsStruct() || field.IsStructPtr()) && !field.IsZero() {
			s, _ := Struct(field.Interface())
			d := s.ToMap(omitZero, omitEmpty)

			// Add embedded fields to the main data.
			if f, _ := r.typ.FieldByName(name); f.Anonymous {
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