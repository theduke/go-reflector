package reflector

import (
	"errors"
	"reflect"
)

type StructReflector struct {
	item       *Reflector
	structItem *Reflector
	isPtr      bool
}

// Struct builds a new StructReflector.
// You may pass in a struct or a pointer to a struct.
func newStructReflector(v *Reflector) (*StructReflector, error) {
	if !v.IsValid() {
		return nil, errors.New(ERR_INVALID_VALUE)
	}

	// Dereference interfaces.
	if v.IsInterface() {
		if v.IsNil() {
			return nil, errors.New(ERR_INVALID_VALUE)
		}
		v = v.Elem()
	}

	if v.IsStruct() {
		return &StructReflector{
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

		return &StructReflector{
			item:       v,
			structItem: v.Elem(),
			isPtr:      true,
		}, nil
	}
	return nil, errors.New(ERR_NOT_A_STRUCT)
}

func (r *StructReflector) Interface() interface{} {
	return r.structItem.Interface()
}

func (r *StructReflector) Value() *Reflector {
	return r.item
}

func (r *StructReflector) Addr() *Reflector {
	if r.item.IsPtr() {
		return r.item
	}
	return r.structItem.Addr()
}

func (r *StructReflector) AddrInterface() interface{} {
	return r.Addr().Interface()
}

func (r *StructReflector) Type() reflect.Type {
	return r.structItem.Type()
}

func (r *StructReflector) Name() string {
	return r.Type().Name()
}

func (r *StructReflector) FullName() string {
	name := r.Name()
	pkg := r.Type().PkgPath()
	if pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func (r *StructReflector) FieldInfo() map[string]*reflect.StructField {
	m := make(map[string]*reflect.StructField, 0)
	for i := 0; i < r.Type().NumField(); i++ {
		f := r.Type().Field(i)
		m[f.Name] = &f
	}
	return m
}

func (r *StructReflector) New() *StructReflector {
	ptr := New(r.structItem.Type())
	refl, err := newStructReflector(ptr)

	// This should never happen, excpet when out of memory!
	// Panic is there just to make sure.
	if err != nil {
		panic(err)
	}

	return refl
}

func (r *StructReflector) Field(fieldName string) *Reflector {
	if !r.HasField(fieldName) {
		return nil
	}
	field := r.structItem.Value().FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	return Reflect(field)
}

func (r *StructReflector) Fields() map[string]*Reflector {
	m := make(map[string]*Reflector)
	for name, _ := range r.FieldInfo() {
		m[name] = Reflect(r.structItem.Value().FieldByName(name))
	}
	return m
}

func (r *StructReflector) EmbeddedFields() map[string]*StructReflector {
	m := make(map[string]*StructReflector)
	for name, field := range r.FieldInfo() {
		if field.Anonymous {
			m[name] = Reflect(r.structItem.Value().FieldByName(name)).MustStruct()
		}
	}
	return m
}

func (r *StructReflector) HasField(fieldName string) bool {
	_, ok := r.Type().FieldByName(fieldName)
	return ok
}

func (r *StructReflector) FieldValue(fieldName string) (interface{}, error) {
	if !r.HasField(fieldName) {
		return nil, errors.New(ERR_UNKNOWN_FIELD)
	}

	field := r.structItem.Value().FieldByName(fieldName)
	if !field.IsValid() {
		return nil, errors.New(ERR_INVALID_FIELD)
	}
	if !field.CanInterface() {
		return nil, errors.New(ERR_UNINTERFACEABLE_FIELD)
	}
	return field.Interface(), nil
}

func (r *StructReflector) UFieldValue(fieldName string) interface{} {
	v, err := r.FieldValue(fieldName)
	if err != nil {
		return nil
	}
	return v
}

func (r *StructReflector) SetFieldValue(fieldName string, value interface{}, convert ...bool) error {
	v := Reflect(value)
	if v == nil {
		return errors.New(ERR_INVALID_VALUE)
	}
	return r.SetField(fieldName, v, convert...)
}

func (r *StructReflector) SetField(fieldName string, value *Reflector, convert ...bool) error {
	field := r.Field(fieldName)
	if field == nil {
		return errors.New(ERR_UNKNOWN_FIELD)
	}
	return field.Set(value, convert...)
}

func (r *StructReflector) ToMap(omitZero, omitEmpty bool) map[string]interface{} {
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

func (r *StructReflector) FromMap(data map[string]interface{}, convert ...bool) error {
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
