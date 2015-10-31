# go-reflector

Go reflector is a library that makes working with reflection in Go easier and safer.

Features:

* Safe methods that do not panic, but return errors or nil
* .IsNumeric(), .IsNil(), .IsZero(), .IsEmpty(), ...
* Easily convert between different types.
* Easily create and work with slices.
* Easily create and work with structs.
* Compare arbitrary values with operators (=, !=, <, <=, >, >=)
* Recursive .ToMap() and .FromMap() for structs
* Sort an array of structs by a struct field

One principal of the library is to almost never panic, but return nil values or errors instead, 
unlike the reflect package of the standard library.
Note that it may panic when using methods prefixed with **Must**.

## Api Documenation

Detailed docs on all methods can be found on [GoDoc](https://godoc.org/github.com/theduke/go-reflector)

## Install

```bash
go get github.com/theduke/go-reflector
```

## Usage

### Basics

```go
import "github.com/theduke/go-reflector"

reflector.Reflect(nil) // => nil!

r := reflector.Reflect(0)
r.IsNumeric() // => true
r.IsNil() // => false
r.IsZero() // => true
r.Len() // => 0 (not an iterable!)

// Convert to another type.
r.ConvertTo("") // => "0"

// Convert to float64.
val, err := r.ConvertTo(float64(0)) // => 0.0

// Convert string to int.
val, err = reflector.Reflect("22").ConvertTo(0) // => 22

// Convert float64 to int.
val, err = reflector.Reflect(float64(10.0)).ConvertTo(0) // => 10

// Convert string to time.Time!
val, err := reflector.Reflect("2012-05-23T18:30:00.000-05:00").ConvertTo(time.Time{}) // => time.Time{}


// Iterables.

r := reflector.Reflect([]int{1,2,3})
r.IsEmpty() // => false
r.IsMap() // => false
r.IsSlice() // => true
r.Len() // => 3
r.Equals(22) // => false
r.Equals([]int{1,2,3}) // => true
```

### Working with slices.

```go
import "github.com/theduke/go-reflector"

s := reflector.Reflect([]int{1,2,3}).MustSlice()
s.Type() // => reflect.Type <int>
s.Len() // => 3
// Get a Reflector for an item in the slice.
s.Index(0) // => Reflector
// Out of range index.
s.Index(22) // => nil
// Get the value at index x
s.IndexValue(0) // => 1
// Out of range index.
s.IndexValue(44) // => nil

// Iterate over all items in the slice.
for _, item := range s.Items() {
	val := item.Interface()
}

// Convert int slice to float slice.
rr, err := s.ConvertTo(float64(0))
floatSlice := rr.Interface() // []float64{1.0, 2.0, 3.0}

// Creating new uint slice.
n := reflector.Reflect(uint(0)).NewSlice()
err := n.AppendValue(uint(4), uint(5), uint(6))
n.Len() // => 3
n.Interface() // => []uint{4, 5, 6}

// When trying to append wrong type.
err := n.AppendValue("2") // => err: type_mismatch

// Updating existing slices.
var intSlice []int
// Note: must pass pointer to slice!
r := reflector.Reflect(&intSlice).MustSlice()
err := r.AppendValue(1, 2)
len(intSlice) // => 2
```

### Working with structs

```go
type TestStruct struct {
	Field1 int
	Field2 string 
}

t := TestStruct{
	Field1: 1, 
	Field2: "x"
}

r := reflector.Reflect(t).MustStruct()

// Check field presence.
r.HasField("Field1") // => true
r.HasField("X") // => false

// Get field values.
val, err := r.FieldValue("Field1") // => interface{}(1), nil
val, err = r.FieldValue("X") // => nil, err_inexistant_field

// Get field values with no error.
r.UFieldValue("Field1") // => interface{}(1), nil
r.UFieldValue("X") // => nil

// Getting reflected fields.
r.Field("Field1").Type() // => reflect.Type <int>
r.Field("Field1").IsZero() // => false
r.Field("Field1").Interface() // => interface{}(1)

// Iterating over all fields:
for name, field := range r.Fields() {
	if field.IsString() {
		...
	}
}

// Recursively (!) convert struct to map.
r.ToMap() // => map[string]interface{}{"Field1": 1, "Field2": "x"}

// Updating field values.

t := TestStruct{
	Field1: 1, 
	Field2: "x"
}

// Notice that you must pass a pointer!
r := reflector.Reflect(&t).MustStruct()

err := r.SetFieldValue("Field1", 44) // => nil
err := r.SetFieldValue("Field2", 44) // => err_type_mismatch
err := r.SetFieldValue("FieldXX", "x") // => err_unknown_field

// Auto-convert values

// Auto-convert from string to int.
// Pass true for auto-convert.
err := r.SetFieldValue("Field1", "44", true) // => nil

// Impossible conversion.
err := r.SetFieldValue("Field1", []int{22}, true) // => err_unconvertable_type

// Load fields from a map.

data := map[string]interface{
	"Field1": float64(0),
	"Field2": "string",
}

// Pass true for auto-convert.
err := r.FromMap(data, true) // => nil
```

### Comparing values

```go

r := reflector.Reflect(20)

flag, err := r.CompareTo(float64(20), "=") // => true, nil
flag, err := r.CompareTo(uint(30), "<") // => true, nil
flag, err := r.CompareTo("20", "!=") // => false, nil

// Invalid comparisons.
flag, err := r.CompareTo([]int{}, "=") // => false, err_incomparable_types
```

### Sort a slice of structs by struct field

```go
type testStruct struct {
	Int int
}

s := []testStruct{
	testStruct{Int: 87},
	testStruct{Int: 1000},
	testStruct{Int: 5},
	testStruct{Int: 800},
	testStruct{Int: 2},
}

r := Reflect(s).MustSlice()
err := SortStructSlice(r, "Int", true)
```

## Additional information

### Changelog

[Changelog](https://github.com/theduke/go-reflector/blob/master/CHANGELOG.md)

### Versioning

This project follows [SemVer](http://semver.org/).

### License

This project is under the [MIT license](https://opensource.org/licenses/MIT).

### Tests

Tests are written in [Ginkgo]()
Test coverage is pretty good, but not perfect (~90%).

To run tests yourself:

```bash
go get github.com/onsi/ginkgo/ginkgo  # installs the ginkgo CLI
go get github.com/onsi/gomega         # fetches the matcher library

cd /path/to/go/reflector
go test -cover
```
