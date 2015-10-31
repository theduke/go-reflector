package reflector_test

import (
	"reflect"
	"time"

	. "github.com/theduke/go-reflector"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testStruct struct {
	Int int
}

type testInterface interface {
}

var _ = Describe("Reflector", func() {

	Describe("IsNumericKind", func() {
		It("Should determine numeric kinds", func() {
			Expect(IsNumericKind(reflect.Int)).To(BeTrue())
		})

		It("Should determine non-numeric kinds", func() {
			Expect(IsNumericKind(reflect.Array)).To(BeFalse())
		})
	})

	Describe("Reflector", func() {
		It("Should return nil for invalid values on Reflect()", func() {
			Expect(Reflect(nil)).To(BeNil())
		})

		It("Should create reflector with Reflect()", func() {
			Expect(Reflect(22)).ToNot(BeNil())
		})

		It("Should return nil on ReflectVal() with invalid val", func() {
			Expect(ReflectVal(reflect.ValueOf(nil))).To(BeNil())
		})

		It("Should create reflector with ReflectVal()", func() {
			v := reflect.ValueOf(22)
			r := ReflectVal(v)
			Expect(r).ToNot(BeNil())
			Expect(r.Value()).To(Equal(v))
		})

		It("Should return interface", func() {
			r := Reflect(22)
			Expect(r.Interface()).To(Equal(22))
		})

		It("Should return value", func() {
			v := 22
			r := Reflect(v)
			Expect(reflect.TypeOf(r.Value())).To(Equal(reflect.TypeOf(reflect.ValueOf(v))))
		})

		It("Should return type", func() {
			Expect(Reflect(22).Type()).To(Equal(reflect.TypeOf(22)))
		})

		It("Should return elem or nil on .Elem()", func() {
			var s *testStruct
			var i testInterface
			Expect(Reflect(0).Elem()).To(BeNil())
			Expect(Reflect(s).Elem()).To(BeNil())
			Expect(Reflect(i)).To(BeNil())

			s = &testStruct{}
			i = &testStruct{}
			Expect(Reflect(s).Elem().Interface()).To(Equal(testStruct{}))
			Expect(Reflect(i).Elem().Interface()).To(Equal(testStruct{}))
		})

		It("Should return nil or Reflector on .Addr()", func() {
			Expect(Reflect(22).Addr()).To(BeNil())
			s := testStruct{
				Int: 22,
			}
			v := reflect.ValueOf(&s).Elem().FieldByName("Int")

			Expect(ReflectVal(v).Addr().Interface()).To(Equal(v.Addr().Interface()))
		})

		It("Should .IsPtr()", func() {
			var x *int
			Expect(Reflect(x).IsPtr()).To(BeTrue())
			Expect(Reflect(22).IsPtr()).To(BeFalse())
		})

		It("Should .IsString()", func() {
			Expect(Reflect("").IsString()).To(BeTrue())
			Expect(Reflect(22).IsString()).To(BeFalse())
		})

		It("Should .IsSlice()", func() {
			Expect(Reflect([]int{22}).IsSlice()).To(BeTrue())
			Expect(Reflect(22).IsSlice()).To(BeFalse())
		})

		It("Should .IsMap()", func() {
			Expect(Reflect(map[string]bool{}).IsMap()).To(BeTrue())
			Expect(Reflect(22).IsMap()).To(BeFalse())
		})

		It("Should .IsStruct()", func() {
			Expect(Reflect(testStruct{}).IsStruct()).To(BeTrue())
			Expect(Reflect(22).IsStruct()).To(BeFalse())
		})

		It("Should .IsStructPtr()", func() {
			var x *testStruct
			Expect(Reflect(x).IsStructPtr()).To(BeTrue())
			Expect(Reflect(22).IsStructPtr()).To(BeFalse())
		})

		/*
			It("Should .IsInterface()", func() {
				var x testInterface = &testStruct{}
				Expect(Reflect(testInterface).IsInterface()).To(BeTrue())
			})
		*/

		It("Should .IsChan()", func() {
			x := make(chan bool)
			Expect(Reflect(x).IsChan()).To(BeTrue())
			Expect(Reflect(22).IsChan()).To(BeFalse())
		})

		It("Should .IsFunc()", func() {
			x := func() {

			}
			Expect(Reflect(x).IsFunc()).To(BeTrue())
			Expect(Reflect(22).IsFunc()).To(BeFalse())
		})

		It("Should .IsArray()", func() {
			var x [5]int
			Expect(Reflect(x).IsArray()).To(BeTrue())
			Expect(Reflect(22).IsArray()).To(BeFalse())
		})

		It("Should .IsBool()", func() {
			var x bool
			Expect(Reflect(x).IsBool()).To(BeTrue())
			Expect(Reflect(22).IsBool()).To(BeFalse())
		})

		It("Should detect numbers with .IsNumeric()", func() {
			Expect(Reflect(int(22)).IsNumeric()).To(BeTrue())
			Expect(Reflect(uint(22)).IsNumeric()).To(BeTrue())
			Expect(Reflect(22.0).IsNumeric()).To(BeTrue())
			Expect(Reflect(float64(22)).IsNumeric()).To(BeTrue())

			Expect(Reflect("").IsNumeric()).To(BeFalse())
		})

		It("Should detect nil values with .IsNil()", func() {
			var ptr *testStruct
			var slice []string
			Expect(Reflect(ptr).IsNil()).To(BeTrue())
			Expect(Reflect(slice).IsNil()).To(BeTrue())

			ptr = &testStruct{}
			Expect(Reflect(ptr).IsNil()).To(BeFalse())
			Expect(Reflect(22).IsNil()).To(BeFalse())
		})

		It("Should detect zero values with .IsZero()", func() {
			var ptr *testStruct
			Expect(Reflect(ptr).IsZero()).To(BeTrue())
			Expect(Reflect(&testStruct{}).IsZero()).To(BeFalse())

			Expect(Reflect(0).IsZero()).To(BeTrue())
			Expect(Reflect(1).IsZero()).To(BeFalse())

			Expect(Reflect(0.0).IsZero()).To(BeTrue())
			Expect(Reflect(0.1).IsZero()).To(BeFalse())

			Expect(Reflect("").IsZero()).To(BeTrue())
			Expect(Reflect("a").IsZero()).To(BeFalse())

			Expect(Reflect(testStruct{}).IsZero()).To(BeTrue())
			Expect(Reflect(testStruct{
				Int: 1,
			}).IsZero()).To(BeFalse())
		})

		It("Should detect zero values with .DeepIsZero()", func() {
			Expect(Reflect(0).DeepIsZero()).To(BeTrue())
			Expect(Reflect(1).DeepIsZero()).To(BeFalse())

			var ptr *testStruct
			Expect(Reflect(ptr).DeepIsZero()).To(BeTrue())
			ptr = &testStruct{}
			Expect(Reflect(ptr).DeepIsZero()).To(BeTrue())
			ptr.Int = 22
			Expect(Reflect(ptr).DeepIsZero()).To(BeFalse())

			var i testInterface
			Expect(Reflect(i)).To(BeNil())
			i = testStruct{}
			Expect(Reflect(i).DeepIsZero()).To(BeTrue())
			i = testStruct{Int: 22}
			Expect(Reflect(i).DeepIsZero()).To(BeFalse())

			// Test nesting.
			var nested testInterface = &testStruct{}
			Expect(Reflect(nested).DeepIsZero()).To(BeTrue())
			nested = &testStruct{Int: 22}
			Expect(Reflect(nested).DeepIsZero()).To(BeFalse())
		})

		It("Should detect empty values with .IsEmpty()", func() {
			var ptr *testStruct
			Expect(Reflect(ptr).IsEmpty()).To(BeTrue())
			Expect(Reflect(&testStruct{}).IsEmpty()).To(BeFalse())
			Expect(Reflect(0).IsEmpty()).To(BeTrue())
			Expect(Reflect(1).IsEmpty()).To(BeFalse())

			// Test slice.
			Expect(Reflect([]int{}).IsEmpty()).To(BeTrue())
			Expect(Reflect([]int{1}).IsEmpty()).To(BeFalse())

			// Test array.
			Expect(Reflect([5]int{}).IsEmpty()).To(BeFalse())

			// Test map.
			Expect(Reflect(map[string]int{}).IsEmpty()).To(BeTrue())
			Expect(Reflect(map[string]int{"test": 1}).IsEmpty()).To(BeFalse())

			// Test chan.
			c := make(chan bool, 2)
			Expect(Reflect(c).IsEmpty()).To(BeTrue())
			c <- true
			Expect(Reflect(c).IsEmpty()).To(BeFalse())
		})

		It("Should create a new slice with .NewSlice()", func() {
			_, ok := Reflect(1).NewSlice().Interface().([]int)
			Expect(ok).To(BeTrue())
		})

		Describe("Type conversions", func() {
			It("Should convert to same type", func() {
				Expect(Reflect(22).ConvertTo(0)).To(Equal(22))
			})

			It("Should convert regular value to pointer", func() {
				v := 22
				p := &v
				Expect(Reflect(v).ConvertTo(p)).To(Equal(p))
			})

			It("Should convert pointer to regular value", func() {
				v := 22
				p := &v
				Expect(Reflect(p).ConvertTo(0)).To(Equal(22))
			})

			It("Should parse time string to time.Time", func() {
				// Test string to time.Time conversion.
				datestr := "2012-05-23T18:30:00.000-05:00"
				t, _ := time.Parse(time.RFC3339, "2012-05-23T18:30:00.000-05:00")
				Expect(Reflect(datestr).ConvertTo(time.Time{})).To(Equal(t))
			})

			It("Should parse time string into *time.Time", func() {
				// Test string to *time.Time conversion.
				datestr := "2012-05-23T18:30:00.000-05:00"
				t, _ := time.Parse(time.RFC3339, "2012-05-23T18:30:00.000-05:00")
				Expect(Reflect(datestr).ConvertTo(&time.Time{})).To(Equal(&t))
			})

			It("Should convert string to bool", func() {
				t := reflect.TypeOf(true)
				Expect(Reflect("y").ConvertToType(t)).To(Equal(true))
				Expect(Reflect("Y").ConvertToType(t)).To(Equal(true))
				Expect(Reflect("yes").ConvertToType(t)).To(Equal(true))
				Expect(Reflect("1").ConvertToType(t)).To(Equal(true))

				Expect(Reflect("n").ConvertToType(t)).To(Equal(false))
				Expect(Reflect("N").ConvertToType(t)).To(Equal(false))
				Expect(Reflect("no").ConvertToType(t)).To(Equal(false))
				Expect(Reflect("0").ConvertToType(t)).To(Equal(false))
			})

			It("Should convert to string", func() {
				Expect(Reflect(time.Time{}).ConvertTo("")).To(Equal("0001-01-01 00:00:00 +0000 UTC"))
				Expect(Reflect(22).ConvertTo("")).To(Equal("22"))
				Expect(Reflect(22.1).ConvertTo("")).To(Equal("22.1"))
			})

			It("Should convert numeric string to number type", func() {
				Expect(Reflect("20").ConvertTo(0)).To(Equal(20))
				Expect(Reflect("20.56").ConvertTo(0.0)).To(Equal(20.56))
			})

			It("Should convert convertible go types", func() {
				Expect(Reflect(20.0).ConvertTo(0)).To(Equal(20))
			})

			It("Should produce error without panicing", func() {
				_, err := Reflect(1).ConvertTo([]int{})
				Expect(err).To(HaveOccurred())
			})

			It("Should convert values with .ConvertToType()", func() {
				v, err := Reflect(22).ConvertToType(reflect.TypeOf(0.1))
				Expect(err).ToNot(HaveOccurred())
				Expect(v).To(Equal(22.0))
			})
		})

		Describe("SetValue()", func() {
			It("Should produce error with unsettable value", func() {
				Expect(Reflect(20).SetValue(55)).To(HaveOccurred())
			})
		})

	})

	Describe("StructReflector", func() {
		It("Should return error on Struct() with invalid value", func() {
			_, err := Struct(nil)
			Expect(err).To(HaveOccurred())
		})

		It("Should return error on Struct() with non-struct", func() {
			_, err := Struct(22)
			Expect(err).To(HaveOccurred())
		})

		It("Should return error on Struct() with pointer to non-struct", func() {
			v := 20
			p := &v
			_, err := Struct(p)
			Expect(err).To(HaveOccurred())
		})

		It("Should create StructReflector with struct", func() {
			Expect(Struct(testStruct{})).ToNot(BeNil())
		})

		It("Should create StructReflector with pointer to struct", func() {
			p := &testStruct{}
			Expect(Struct(p)).ToNot(BeNil())
		})

		It("Should panic on MustStruct() with non-struct arg", func() {
			f := func() {
				MustStruct(22)
			}
			Expect(f).Should(Panic())
		})

		It("Should create reflector with MustStruct()", func() {
			Expect(MustStruct(testStruct{})).ToNot(BeNil())
		})

		It("Should run .Interface()", func() {
			r, _ := Struct(testStruct{})
			Expect(r.Interface()).To(Equal(testStruct{}))
		})

		It("Should run .Value()", func() {
			s := testStruct{}
			r, _ := Struct(s)
			Expect(r.Value().Interface()).To(Equal(s))
		})

		It("Should run .Type()", func() {
			s := testStruct{}
			r, _ := Struct(s)
			Expect(r.Type()).To(Equal(reflect.TypeOf(s)))
		})

		It("Should create new struct with .New()", func() {
			r, _ := Struct(testStruct{})
			s := r.New()
			Expect(Reflect(s).Elem().Interface()).To(Equal(testStruct{}))
		})

		It("Should return nil on .Field() with inexistant field", func() {
			r := MustStruct(testStruct{})
			Expect(r.Field("InexistantField")).To(BeNil())
		})

		It("Should return Reflector on .Field()", func() {
			r := MustStruct(testStruct{})
			Expect(r.Field("Int")).ToNot(BeNil())
		})

		It("Should return true on .HasField() with valid field", func() {
			r := MustStruct(testStruct{})
			Expect(r.HasField("Int")).To(BeTrue())
		})

		It("Should return false on .HasField() with inexistant field", func() {
			r := MustStruct(testStruct{})
			Expect(r.HasField("InexistantField")).To(BeFalse())
		})

		It("Should return error on .FieldValue() with inexistant field", func() {
			r := MustStruct(testStruct{})
			_, err := r.FieldValue("InexistantField")
			Expect(err).To(HaveOccurred())
		})

		It("Should return val .FieldValue()", func() {
			r := MustStruct(testStruct{Int: 22})
			Expect(r.FieldValue("Int")).To(Equal(22))
		})

		It("Should return nil on .UFieldValue() with inexistant field", func() {
			r := MustStruct(testStruct{Int: 22})
			Expect(r.UFieldValue("InexistantField")).To(BeNil())
		})

		It("Should return value on .UFieldValue()", func() {
			r := MustStruct(testStruct{Int: 22})
			Expect(r.UFieldValue("Int")).To(Equal(22))
		})

		It("Should return value on .UFieldValue()", func() {
			r := MustStruct(testStruct{Int: 22})
			Expect(r.UFieldValue("Int")).To(Equal(22))
		})

		It("Should return error on .SetField() with inexistant field", func() {
			s := &testStruct{}
			r := MustStruct(s)
			Expect(r.SetField("InexistantField", 22)).To(HaveOccurred())
		})

		It("Should .SetField()", func() {
			s := &testStruct{}
			r := MustStruct(s)
			Expect(r.SetField("Int", 22)).ToNot(HaveOccurred())
			Expect(s.Int).To(Equal(22))
		})

		It("Should fail .SetField() with type mismatch", func() {
			s := &testStruct{}
			r := MustStruct(s)
			Expect(r.SetField("Int", 22.0)).To(HaveOccurred())
		})

		It("Should .SetField() with non-matching types and conversion enabled", func() {
			s := &testStruct{}
			r := MustStruct(s)
			Expect(r.SetField("Int", 22.0, true)).ToNot(HaveOccurred())
		})

		It("Should fail .SetField() with conversion enabled and inconvertable types", func() {
			s := &testStruct{}
			r := MustStruct(s)
			Expect(r.SetField("Int", []int{})).To(HaveOccurred())
		})
	})
})

/**

// IsNumeric returns true if the value is any numeric type (uint, int, float64, ...)
IsNumeric() bool

*/