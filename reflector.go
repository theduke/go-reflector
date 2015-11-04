// go-reflector is a utility library that makes working with reflection easier and safer.
//
// For more information an usage examples, check the Github repository at
// https://github.com/theduke/go-reflector.
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
	ERR_NOT_A_SLICE                = "not_a_slice"
	ERR_UNSETTABLE_VALUE           = "unsettable_value"
	ERR_TYPE_MISMATCH              = "type_mismatch"
	ERR_UNCOMPARABLE_VALUES        = "uncomparable_values"
	ERR_UNKNOWN_OPERATOR           = "unknown_operator"
	ERR_CANT_APPEND_NOT_A_POINTER  = "cant_append_when_slice_reflector_not_created_from_pointer"
	ERR_INDEX_OUT_OF_BOUNDS        = "index_out_of_bounds"
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

// New creates a new instance of the given type, and returns a Reflector.
func New(typ reflect.Type) *Reflector {
	return Reflect(reflect.New(typ))
}

// Reflect returns a new Reflector for the given value, or nil if the value is invalid.
func Reflect(value interface{}) *Reflector {
	var val reflect.Value
	if v, ok := value.(reflect.Value); ok {
		val = v
	} else {
		val = reflect.ValueOf(value)
	}

	r := &Reflector{
		value: val,
	}
	if r.IsInterface() {
		r = r.Elem()
	}
	return r
}

// R is a convenient alias for Reflect().
func R(value interface{}) *Reflector {
	return Reflect(value)
}

type Reflector struct {
	value reflect.Value
}

func (r *Reflector) String() string {
	return r.value.String()
}

func (r *Reflector) Interface() interface{} {
	val := r.value
	if r.IsInterface() {
		val = r.Elem().Value()
	}

	if !val.CanInterface() {
		return nil
	}
	return val.Interface()
}

func (r *Reflector) Value() reflect.Value {
	return r.value
}

func (r *Reflector) Type() reflect.Type {
	return r.value.Type()
}

func (r *Reflector) Kind() reflect.Kind {
	return r.Type().Kind()
}

func (r *Reflector) Elem() *Reflector {
	if r.IsNil() {
		return nil
	}
	if !(r.IsPtr() || r.IsInterface()) {
		return nil
	}
	return Reflect(r.value.Elem())
}

func (r *Reflector) Addr() *Reflector {
	if !r.value.CanAddr() {
		return nil
	}
	return Reflect(r.value.Addr().Interface())
}

func (r *Reflector) New() *Reflector {
	return New(r.Type())
}

func (r *Reflector) IsPtr() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Ptr
}

func (r *Reflector) IsString() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.String
}

func (r *Reflector) IsSlice() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Slice
}

func (r *Reflector) IsMap() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Map
}

func (r *Reflector) IsStruct() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Struct
}

func (r *Reflector) IsStructPtr() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Ptr && r.Type().Elem().Kind() == reflect.Struct
}

func (r *Reflector) IsInterface() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Interface
}

func (r *Reflector) IsChan() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Chan
}

func (r *Reflector) IsFunc() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Func
}

func (r *Reflector) IsArray() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Array
}

func (r *Reflector) IsBool() bool {
	if !r.IsValid() {
		return false
	}
	return r.Type().Kind() == reflect.Bool
}

func (r *Reflector) IsNumeric() bool {
	if !r.IsValid() {
		return false
	}
	return IsNumericKind(r.Type().Kind())
}

func (r *Reflector) IsIterable() bool {
	switch r.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return true
	}
	return false
}

func (r *Reflector) Len() int {
	if r.IsIterable() {
		return r.value.Len()
	}
	return 0
}

func (r *Reflector) IsValid() bool {
	return r.value.IsValid()
}

func (r *Reflector) IsNil() bool {
	if !r.IsValid() {
		return true
	}
	switch r.Type().Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		// Only these types can be nil.
		return r.value.IsNil()
	}
	// Not a nillable type, so just return false.
	return false
}

func (r *Reflector) IsZero() bool {
	if r.IsNil() {
		return true
	}
	// Not nil, so compare with the zero type.

	// Prevent comparison with uncomparable types.
	if r.IsSlice() || r.IsArray() || r.IsMap() {
		return false
	}
	return r.Interface() == reflect.Zero(r.Type()).Interface()
}

func (r *Reflector) DeepIsZero() bool {
	if r.IsZero() {
		return true
	}
	if r.IsPtr() || r.IsInterface() {
		return r.Elem().DeepIsZero()
	}
	return false
}

func (r *Reflector) IsEmpty() bool {
	if r.IsZero() {
		return true
	}

	switch r.Type().Kind() {
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Chan:
		return r.value.Len() < 1
	}

	return false
}

func (r *Reflector) Equals(value interface{}) bool {
	// De-reference Reflectors or reflect.Value.
	if r, ok := value.(*Reflector); ok {
		value = r.Interface()
	} else if v, ok := value.(reflect.Value); ok {
		value = v.Interface()
	}
	return reflect.DeepEqual(r.Interface(), value)
}

func (r *Reflector) Slice() (*SliceReflector, error) {
	return newSliceReflector(r)
}

func (r *Reflector) MustSlice() *SliceReflector {
	s, err := r.Slice()
	if err != nil {
		panic(err)
	}
	return s
}

func (r *Reflector) Struct() (*StructReflector, error) {
	return newStructReflector(r)
}

func (r *Reflector) MustStruct() *StructReflector {
	s, err := r.Struct()
	if err != nil {
		panic(err)
	}
	return s
}

func (r *Reflector) NewSlice() *SliceReflector {
	// Build new array.
	// See http://stackoverflow.com/questions/25384640/why-golang-reflect-makeslice-returns-un-addressable-value
	// Create a slice to begin with
	s := reflect.MakeSlice(reflect.SliceOf(r.Type()), 0, 0)

	// Create a pointer to a slice value and set it to the slice
	x := reflect.New(s.Type())
	x.Elem().Set(s)

	sliceReflector, err := newSliceReflector(Reflect(x))
	if err != nil {
		// This should never happen!
		// Panic just to be sure, though.
		panic(err)
	}
	return sliceReflector
}

func (r *Reflector) ConvertTo(targetVal interface{}) (interface{}, error) {
	if typ, ok := targetVal.(reflect.Type); ok {
		return r.ConvertToType(typ)
	}

	// Check for empty value, to prevent a panic when user
	// passes in nil for example.
	if !reflect.ValueOf(targetVal).IsValid() {
		return nil, errors.New(ERR_INVALID_VALUE)
	}
	return r.ConvertToType(reflect.TypeOf(targetVal))
}

func (r *Reflector) saveConvertToType(typ reflect.Type) interface{} {
	defer func() {
		recover()
	}()

	return r.value.Convert(typ).Interface()
}

func (r *Reflector) ConvertToType(typ reflect.Type) (interface{}, error) {
	kind := typ.Kind()

	valKind := r.Type().Kind()

	if typ == r.Type() {
		// Same type, nothing to convert.
		return r.Interface(), nil
	}

	// If value and target are slices, but of differing type, try to convert.
	if r.IsSlice() && typ.Kind() == reflect.Slice {
		sliceR, err := r.Slice()
		if err != nil {
			return nil, err
		}
		newSlice, err := sliceR.ConvertToType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return newSlice, nil
	}

	isPointer := kind == reflect.Ptr
	var pointerType reflect.Type
	if isPointer {
		pointerType = typ.Elem()
	}

	// If target value is a pointer and the value is not (and the types match),
	// create a new pointer pointing to the value.
	if isPointer && r.Type() == pointerType {
		newVal := reflect.New(r.Type())
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
		date, err := time.Parse(time.RFC3339, r.Interface().(string))
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
		str := strings.ToLower(strings.TrimSpace(r.Interface().(string)))
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
		if bytes, ok := r.Interface().([]byte); ok {
			return string(bytes), nil
		}

		// Check if type implemens stringer interface.
		if stringer, ok := r.Interface().(fmt.Stringer); ok {
			// Implements Stringer, so use .String().
			return stringer.String(), nil
		}

		// Does not implement stringer, so use fmt package.
		return fmt.Sprintf("%v", r.Interface()), nil
	}

	// If value is string, and target type is numeric,
	// parse to float and then convert with reflect.
	if valKind == reflect.String && IsNumericKind(kind) {
		num, err := strconv.ParseFloat(r.Interface().(string), 64)
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

func (r *Reflector) Set(value *Reflector, convert ...bool) error {
	if value == nil {
		return errors.New(ERR_INVALID_VALUE)
	} else if !value.IsValid() {
		return errors.New(ERR_INVALID_VALUE)
	} else if !r.value.CanSet() {
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

func (r *Reflector) SetValue(rawValue interface{}, convert ...bool) error {
	val := Reflect(rawValue)
	if !val.IsValid() {
		return errors.New(ERR_INVALID_VALUE)
	}

	return r.Set(val, convert...)
}

func compareStringValues(condition, a, b string) (bool, error) {
	// Check different possible filters.
	switch condition {
	case "=", "==":
		return a == b, nil
	case "!=":
		return a != b, nil
	case "like":
		return strings.Contains(a, b), nil
	case "<":
		return a < b, nil
	case "<=":
		return a <= b, nil
	case ">":
		return a > b, nil
	case ">=":
		return a >= b, nil
	}

	// Should never happen, since .CompareTo checks the operator.
	panic("Unknown operator: " + condition)
}

func compareFloat64Values(condition string, a, b float64) (bool, error) {
	// Check different possible filters.
	switch condition {
	case "=", "==":
		return a == b, nil
	case "!=":
		return a != b, nil
	case "like":
		return false, errors.New("invalid_filter_comparison: LIKE filter can only be used for string values, not numbers")
	case "<":
		return a < b, nil
	case "<=":
		return a <= b, nil
	case ">":
		return a > b, nil
	case ">=":
		return a >= b, nil
	}

	// Should never happen, since .CompareTo checks the operator.
	panic("Unknown operator: " + condition)
}

func (r *Reflector) CompareTo(value interface{}, operator string) (bool, error) {
	// Check operator.
	switch operator {
	case "=", "!=", "like", "<", "<=", ">", ">=":
	case "==":
		operator = "="
	default:
		return false, errors.New(ERR_UNKNOWN_OPERATOR)
	}

	a := r
	aVal := r.Interface()
	if !a.IsValid() || a.DeepIsZero() {
		aVal = float64(0)
		a = Reflect(aVal)
	}
	if a.IsPtr() {
		a = a.Elem()
		aVal = a.Interface()
	}
	typA := a.Type()
	kindA := typA.Kind()

	bVal := value
	var b *Reflector
	if r, ok := value.(*Reflector); ok {
		// Value is a reflector, so use it.
		b = r
		bVal = b.Interface()
	} else {
		b = Reflect(value)
	}
	if !b.IsValid() || b.DeepIsZero() {
		bVal = float64(0)
		b = Reflect(bVal)
		aVal, bVal = bVal, aVal
		a, b = b, a
	}
	if b.IsPtr() {
		b = b.Elem()
		bVal = b.Interface()
	}
	typB := b.Type()
	kindB := typB.Kind()

	// Compare time.Time values numerically.
	if kindA == reflect.Struct && typA.PkgPath() == "time" && typA.Name() == "Time" {
		t := aVal.(time.Time)
		aVal = float64(t.UnixNano())
		a = Reflect(aVal)
		typA = a.Type()
		kindA = typA.Kind()
	}

	if kindB == reflect.Struct && typB.PkgPath() == "time" && typB.Name() == "Time" {
		t := bVal.(time.Time)
		bVal = float64(t.UnixNano())
		b = Reflect(bVal)
		typB = b.Type()
		kindB = typB.Kind()
	}

	// Convert time.Duration to float64.
	if d, ok := aVal.(time.Duration); ok {
		aVal = float64(d)
		a = Reflect(aVal)
		typA = a.Type()
		kindA = typA.Kind()
	}
	if d, ok := bVal.(time.Duration); ok {
		bVal = float64(d)
		b = Reflect(aVal)
		typB = b.Type()
		kindB = typB.Kind()
	}

	if IsNumericKind(kindA) || IsNumericKind(kindB) {
		numA, err := a.ConvertTo(float64(0))
		if err != nil {
			return false, errors.New("Conversion error: " + err.Error())
		}

		numB, err := b.ConvertTo(float64(0))
		if err != nil {
			return false, errors.New("Conversion error: " + err.Error())
		}

		return compareFloat64Values(operator, numA.(float64), numB.(float64))
	}

	if kindA == reflect.String {
		convertedB, err := b.ConvertTo("")
		if err != nil {
			return false, errors.New("Conversion error: " + err.Error())
		}
		return compareStringValues(operator, aVal.(string), convertedB.(string))
	}

	if operator == "=" || operator == "!=" {
		convertedB, err := b.ConvertToType(typA)
		if err != nil {
			return false, errors.New("Conversion error: " + err.Error())
		}

		if operator == "=" {
			return a.Equals(convertedB), nil
		} else {
			return !a.Equals(convertedB), nil
		}
	}

	msg := "impossible_comparison: " + fmt.Sprintf("Cannot compare type %v(value %v) to type %v(value %v)", kindA, aVal, kindB, bVal)
	return false, errors.New(msg)
}
