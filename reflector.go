/**
 * reflector is a utility library that makes working with reflection easier.
 *
 * For more information an usage examples, check the github repository at
 * https://github.com/theduke/go-reflector.
 */
package reflector

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	ERR_UNKNOWN_FIELD         = "unknown_field"
	ERR_INVALID_FIELD         = "invalid_field"
	ERR_UNINTERFACEABLE_FIELD = "uninterfaceable_field"

	ERR_INVALID_TIME        = "invalid_time_not_rfc3339"
	ERR_UNCONVERTABLE_TYPES = "unconvertable_types"

	ERR_POINTER_OR_STRUCT_EXPECTED = "pointer_or_struct_expected"
	ERR_INVALID_VALUE              = "invalid_value"
	ERR_STRUCT_EXPECTED            = "struct_expected"
	ERR_NIL_POINTER                = "nil_pointer"
	ERR_NOT_A_STRUCT               = "not_a_struct"
	ERR_UNSETTABLE_VALUE           = "unsettable_value"
	ERR_TYPE_MISMATCH              = "type_mismatch"
)

// IsNumericKind returns true if the given reflect.Kind is any numeric type,
// like int, uint32, ...
func IsNumericKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func New(typ reflect.Type) Reflector {
	return ReflectVal(reflect.New(typ))
}

type Reflector interface {
	// Interface returns the value as interface{} or nil if the value can't be
	// interfaced.
	//
	// Note that in contrast to reflect.Value.Interface(), this method can not
	// panic, but returns nil instead.
	Interface() interface{}

	// Value returns the raw reflect.Value.
	Value() reflect.Value

	// Type returns the raw reflect.Type.
	Type() reflect.Type

	// Elem will return the value a pointer points to, or the value an
	// interface contains.
	// If the value is neither an interface or a pointer, or the
	// pointer/interface is nil, it will return nil.
	Elem() Reflector

	// Addr will return a Reflector for the address of the value, or nil if it
	// can not be addressed.
	Addr() Reflector

	IsPtr() bool
	IsString() bool
	IsSlice() bool
	IsMap() bool
	IsStruct() bool
	// IsStructPtr returns true if the value is a pointer holding a struct.
	IsStructPtr() bool
	IsInterface() bool
	IsChan() bool
	IsFunc() bool
	IsArray() bool
	IsBool() bool

	// IsNumeric returns true if the value is any numeric type (uint, int, float64, ...)
	IsNumeric() bool

	// IsNil returns true if the value has a nil-able type and is nil, or false.
	//
	// Note that in constrast to reflect.Value.IsNil() this method can't panic,
	// and will just return false for values that can't be nil.
	IsNil() bool

	// IsZero returns true if the value is nil (checked with Reflector.IsNil())
	// or equal to the zero value of the type.
	IsZero() bool

	// DeepIsZero is mostly the same as IsZero, but if a pointer is given, it
	// will dereference the pointer and will check if the value pointed to is
	// zero.
	DeepIsZero() bool

	// IsEmpty returns true if the value is empty.
	//
	// It first checks if the value is zero with Reflector.IsZero().
	// If the value is not zero, it will return true for:
	// * empty maps.
	// * slices of length 0.
	// * arrays of length 0.
	// * strings of length 0.
	// * chans with 0 items.
	//
	// If non of these tests match, it will return false.
	IsEmpty() bool

	// Struct returns a new StructReflector if the value is either a struct or a
	// pointer to a struct. Returns nil and an error otherwise.
	Struct() (StructReflector, error)

	// Creates a new slice holding the same type as the value.
	// Then returns a pointer to the slice as a reflect.Value.
	NewSlice() reflect.Value

	// ConvertTo tries to convert the value to the same type as the passed in
	// value.
	// If successful, the converted value is returned, or nil and an error
	// otherwise.
	//
	// For details on conversion rules, see ConvertToType().
	//
	// To convert to a reflect.Type, use Reflector.ConvertToType.
	// If successful, the converted value is returned, or nil and an error
	// otherwise.
	//
	// Can not panic!
	ConvertTo(typeValue interface{}) (interface{}, error)

	// ConvertToType tries to convert the value to the given type.
	// Returns the converted value if successful, or nil and an error.
	//
	// Can not panic!
	ConvertToType(targetType reflect.Type) (interface{}, error)

	// Sets the field to the given value.
	//
	// Returns an error if the types are not compatible.
	// If you pass true as a second argument, a type conversion will be
	// attempted.
	Set(value Reflector, convert ...bool) error

	// Sets the field to the given value.
	//
	// Returns an error if the types are not compatible.
	// If you pass true as a second argument, a type conversion will be
	// attempted.
	SetValue(value interface{}, convert ...bool) error
}

// Reflect returns a new Reflector for the given value, or nil if the value
// is invalid.
func Reflect(value interface{}) Reflector {
	val := reflect.ValueOf(value)
	if !val.IsValid() {
		return nil
	}

	return &reflector{
		rawValue: value,
		value:    val,
		typ:      val.Type(),
	}
}

func ReflectVal(val reflect.Value) Reflector {
	if !val.IsValid() {
		return nil
	}
	return &reflector{
		rawValue: val.Interface(),
		value:    val,
		typ:      val.Type(),
	}
}

type reflector struct {
	rawValue interface{}
	value    reflect.Value
	typ      reflect.Type
}

// Ensure reflector implements Reflector.
var _ Reflector = (*reflector)(nil)

func (r *reflector) Interface() interface{} {
	if !r.value.CanInterface() {
		return nil
	}
	return r.value.Interface()
}

func (r *reflector) Value() reflect.Value {
	return r.value
}

func (r *reflector) Type() reflect.Type {
	return r.typ
}

func (r *reflector) Elem() Reflector {
	if r.IsNil() {
		return nil
	}
	if !(r.IsPtr() || r.IsInterface()) {
		return nil
	}
	return ReflectVal(r.value.Elem())
}

func (r *reflector) Addr() Reflector {
	if !r.value.CanAddr() {
		return nil
	}
	return Reflect(r.value.Addr().Interface())
}

func (r *reflector) IsPtr() bool {
	return r.typ.Kind() == reflect.Ptr
}

func (r *reflector) IsString() bool {
	return r.typ.Kind() == reflect.String
}

func (r *reflector) IsSlice() bool {
	return r.typ.Kind() == reflect.Slice
}

func (r *reflector) IsMap() bool {
	return r.typ.Kind() == reflect.Map
}

func (r *reflector) IsStruct() bool {
	return r.typ.Kind() == reflect.Struct
}

func (r *reflector) IsStructPtr() bool {
	return r.typ.Kind() == reflect.Ptr && r.typ.Elem().Kind() == reflect.Struct
}

func (r *reflector) IsInterface() bool {
	return r.typ.Kind() == reflect.Interface
}

func (r *reflector) IsChan() bool {
	return r.typ.Kind() == reflect.Chan
}

func (r *reflector) IsFunc() bool {
	return r.typ.Kind() == reflect.Func
}

func (r *reflector) IsArray() bool {
	return r.typ.Kind() == reflect.Array
}

func (r *reflector) IsBool() bool {
	return r.typ.Kind() == reflect.Bool
}

func (r *reflector) IsNumeric() bool {
	return IsNumericKind(r.typ.Kind())
}

func (r *reflector) IsNil() bool {
	switch r.typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		// Only these types can be nil.
		return r.value.IsNil()
	}
	// Not a nillable type, so just return false.
	return false
}

func (r *reflector) IsZero() bool {
	if r.IsNil() {
		return true
	}
	// Not nil, so compare with the zero type.

	// Prevent comparison with uncomparable types.
	if r.IsSlice() || r.IsArray() || r.IsMap() {
		return false
	}
	return r.rawValue == reflect.Zero(r.typ).Interface()
}

func (r *reflector) DeepIsZero() bool {
	if r.IsZero() {
		return true
	}
	if r.IsPtr() || r.IsInterface() {
		return r.Elem().DeepIsZero()
	}
	return false
}

func (r *reflector) IsEmpty() bool {
	if r.IsZero() {
		return true
	}

	switch r.typ.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Chan:
		return r.value.Len() < 1
	}

	return false
}

func (r *reflector) Struct() (StructReflector, error) {
	if r.IsStruct() {
		return Struct(r.Addr().Interface())
	} else if r.IsStructPtr() {
		if r.IsNil() {
			newStruct := New(r.Type().Elem())
			r.Set(newStruct)
			return Struct(newStruct.Interface())
		} else {
			return Struct(r.rawValue)
		}
	}
	return nil, errors.New(ERR_NOT_A_STRUCT)
}

func (r *reflector) NewSlice() reflect.Value {
	// Build new array.
	// See http://stackoverflow.com/questions/25384640/why-golang-reflect-makeslice-returns-un-addressable-value
	// Create a slice to begin with

	slice := reflect.MakeSlice(reflect.SliceOf(r.typ), 0, 0)

	// Create a pointer to a slice value and set it to the slice
	x := reflect.New(slice.Type())
	x.Elem().Set(slice)

	return x.Elem()
}

func (r *reflector) ConvertTo(targetVal interface{}) (interface{}, error) {
	// Check for empty value, to prevent a panic when user
	// passes in nil for example.
	if !reflect.ValueOf(targetVal).IsValid() {
		return nil, errors.New(ERR_INVALID_VALUE)
	}
	return r.ConvertToType(reflect.TypeOf(targetVal))
}

func (r *reflector) saveConvertToType(typ reflect.Type) interface{} {
	defer func() {
		recover()
	}()

	return r.value.Convert(typ).Interface()
}

func (r *reflector) ConvertToType(typ reflect.Type) (interface{}, error) {
	kind := typ.Kind()

	valKind := r.typ.Kind()

	if typ == r.typ {
		// Same type, nothing to convert.
		return r.rawValue, nil
	}

	isPointer := kind == reflect.Ptr
	var pointerType reflect.Type
	if isPointer {
		pointerType = typ.Elem()
	}

	// If target value is a pointer and the value is not (and the types match),
	// create a new pointer pointing to the value.
	if isPointer && r.typ == pointerType {
		newVal := reflect.New(r.typ)
		newVal.Elem().Set(r.value)

		return newVal.Interface(), nil
	}

	// If value is a pointer, and the target is not, and the types match,
	// take the elem of the value.
	if r.IsPtr() && !isPointer && r.Type().Elem() == typ {
		return r.Elem().Interface(), nil
	}

	// Parse dates into time.Time.

	isTime := kind == reflect.Struct && typ.PkgPath() == "time" && typ.Name() == "Time"
	isTimePointer := isPointer && pointerType.Kind() == reflect.Struct && pointerType.PkgPath() == "time" && pointerType.Name() == "Time"

	if (isTime || isTimePointer) && valKind == reflect.String {
		date, err := time.Parse(time.RFC3339, r.rawValue.(string))
		if err != nil {
			return nil, errors.New(ERR_INVALID_TIME)
		}

		if isTime {
			return date, nil
		} else {
			return &date, nil
		}
	}

	// Special handling for bool to string.
	if kind == reflect.Bool && r.IsString() {
		str := strings.ToLower(strings.TrimSpace(r.rawValue.(string)))
		switch str {
		case "y", "yes", "1":
			return true, nil
		case "n", "no", "0":
			return false, nil
		}
	}

	// Special handling for string target.
	if kind == reflect.String {
		// Convert byte array to string.
		if bytes, ok := r.rawValue.([]byte); ok {
			return string(bytes), nil
		}

		// Check if type implemens stringer interface.
		if stringer, ok := r.rawValue.(fmt.Stringer); ok {
			// Implements Stringer, so use .String().
			return stringer.String(), nil
		}

		// Does not implement stringer, so use fmt package.
		return fmt.Sprintf("%v", r.rawValue), nil
	}

	// If value is string, and target type is numeric,
	// parse to float and then convert with reflect.
	if valKind == reflect.String && IsNumericKind(kind) {
		num, err := strconv.ParseFloat(r.rawValue.(string), 64)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(num).Convert(typ).Interface(), nil
	}

	// No custom handling worked, so try to convert with reflect.
	converted := r.saveConvertToType(typ)
	if converted == nil {
		return nil, errors.New(ERR_UNCONVERTABLE_TYPES)
	}

	return converted, nil
}

func (r *reflector) Set(value Reflector, convert ...bool) error {
	if !r.value.CanSet() {
		return errors.New(ERR_UNSETTABLE_VALUE)
	}
	doConvert := len(convert) > 0 && convert[0]
	if value.Type() != r.Type() {
		if doConvert {
			// Try to convert.
			converted, err := value.ConvertToType(r.Type())
			if err != nil {
				return err
			}
			value = Reflect(converted)
		} else {
			return errors.New(ERR_TYPE_MISMATCH)
		}
	}
	r.value.Set(value.Value())
	return nil
}

func (r *reflector) SetValue(rawValue interface{}, convert ...bool) error {
	val := Reflect(rawValue)
	if val == nil {
		return errors.New(ERR_INVALID_VALUE)
	}

	return r.Set(val, convert...)
}

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
