package reflector_test

import (
	"reflect"

	. "github.com/theduke/go-reflector"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Struct", func() {
	It("Should return error on Struct() with non-struct", func() {
		_, err := Reflect(22).Struct()
		Expect(err).To(HaveOccurred())
	})

	It("Should return error on Struct() with pointer to non-struct", func() {
		v := 20
		p := &v
		_, err := Reflect(p).Struct()
		Expect(err).To(HaveOccurred())
	})

	It("Should create StructReflector with struct", func() {
		Expect(Reflect(testStruct{}).Struct()).ToNot(BeNil())
	})

	It("Should create StructReflector with pointer to struct", func() {
		p := &testStruct{}
		Expect(Reflect(p).Struct()).ToNot(BeNil())
	})

	It("Should panic on MustStruct() with non-struct arg", func() {
		f := func() {
			Reflect(22).MustStruct()
		}
		Expect(f).Should(Panic())
	})

	It("Should create reflector with MustStruct()", func() {
		Expect(Reflect(testStruct{}).MustStruct()).ToNot(BeNil())
	})

	It("Should run .Interface()", func() {
		r, _ := Reflect(testStruct{}).Struct()
		Expect(r.Interface()).To(Equal(testStruct{}))
	})

	It("Should run .Value()", func() {
		s := Reflect(testStruct{})
		r, _ := s.Struct()
		Expect(r.Value().Interface()).To(Equal(s.Interface()))
	})

	It("Should run .Type()", func() {
		s := testStruct{}
		r, _ := Reflect(s).Struct()
		Expect(r.Type()).To(Equal(reflect.TypeOf(s)))
	})

	It("Should create new struct with .New()", func() {
		r, _ := Reflect(testStruct{}).Struct()
		s := r.New()
		Expect(s.Value().Interface()).To(Equal(testStruct{}))
	})

	It("Should return nil on .Field() with inexistant field", func() {
		r := Reflect(testStruct{}).MustStruct()
		Expect(r.Field("InexistantField")).To(BeNil())
	})

	It("Should return Reflector on .Field()", func() {
		r := Reflect(testStruct{}).MustStruct()
		Expect(r.Field("Int")).ToNot(BeNil())
	})

	It("Should build field map with .Fields()", func() {
		s := testStruct{
			Int:    10,
			Float:  10.10,
			String: "str",
		}
		r := Reflect(s).MustStruct()
		fields := r.Fields()
		Expect(fields).To(HaveKey("Int"))
		Expect(fields).To(HaveKey("Float"))
		Expect(fields).To(HaveKey("String"))
		Expect(fields["Int"].Interface()).To(Equal(10))
	})

	It("Should return true on .HasField() with valid field", func() {
		r := Reflect(testStruct{}).MustStruct()
		Expect(r.HasField("Int")).To(BeTrue())
	})

	It("Should return false on .HasField() with inexistant field", func() {
		r := Reflect(testStruct{}).MustStruct()
		Expect(r.HasField("InexistantField")).To(BeFalse())
	})

	It("Should return error on .FieldValue() with inexistant field", func() {
		r := Reflect(testStruct{}).MustStruct()
		_, err := r.FieldValue("InexistantField")
		Expect(err).To(HaveOccurred())
	})

	It("Should return val .FieldValue()", func() {
		r := Reflect(testStruct{Int: 22}).MustStruct()
		Expect(r.FieldValue("Int")).To(Equal(22))
	})

	It("Should return nil on .UFieldValue() with inexistant field", func() {
		r := Reflect(testStruct{Int: 22}).MustStruct()
		Expect(r.UFieldValue("InexistantField")).To(BeNil())
	})

	It("Should return value on .UFieldValue()", func() {
		r := Reflect(testStruct{Int: 22}).MustStruct()
		Expect(r.UFieldValue("Int")).To(Equal(22))
	})

	It("Should return value on .UFieldValue()", func() {
		r := Reflect(testStruct{Int: 22}).MustStruct()
		Expect(r.UFieldValue("Int")).To(Equal(22))
	})

	It("Should return error on .SetField() with inexistant field", func() {
		s := &testStruct{}
		r := Reflect(s).MustStruct()
		Expect(r.SetFieldValue("InexistantField", 22)).To(HaveOccurred())
	})

	It("Should .SetField()", func() {
		s := &testStruct{}
		r := Reflect(s).MustStruct()
		Expect(r.SetFieldValue("Int", 22)).ToNot(HaveOccurred())
		Expect(s.Int).To(Equal(22))
	})

	It("Should fail .SetField() with type mismatch", func() {
		s := &testStruct{}
		r := Reflect(s).MustStruct()
		Expect(r.SetFieldValue("Int", 22.0)).To(HaveOccurred())
	})

	It("Should .SetField() with non-matching types and conversion enabled", func() {
		s := &testStruct{}
		r := Reflect(s).MustStruct()
		Expect(r.SetFieldValue("Int", 22.0, true)).ToNot(HaveOccurred())
	})

	It("Should fail .SetField() with conversion enabled and inconvertable types", func() {
		s := &testStruct{}
		r := Reflect(s).MustStruct()
		Expect(r.SetFieldValue("Int", []int{})).To(HaveOccurred())
	})

	Describe("Map conversions", func() {
		It("Should convert to map", func() {
			s := nestedStruct{
				testStruct: testStruct{
					Int:    10,
					Float:  10.10,
					String: "str",
				},

				Embedded: testStruct{
					Int:    20,
					Float:  20.20,
					String: "str2",
				},

				EmbeddedPtr: &testStruct{
					Int:    30,
					Float:  30.30,
					String: "str3",
				},
			}

			data := Reflect(s).MustStruct().ToMap(false, true)

			d := map[string]interface{}{
				"Int":    10,
				"Float":  10.10,
				"String": "str",

				"Embedded": map[string]interface{}{
					"Int":    20,
					"Float":  20.20,
					"String": "str2",
				},

				"EmbeddedPtr": map[string]interface{}{
					"Int":    30,
					"Float":  30.30,
					"String": "str3",
				},
			}

			Expect(data).To(Equal(d))
		})

		It("Should ignore zero fields on ToMap", func() {
			s := nestedStruct{
				testStruct: testStruct{
					Int:    10,
					Float:  10.10,
					String: "str",
				},
			}

			data := Reflect(s).MustStruct().ToMap(true, false)

			d := map[string]interface{}{
				"Int":    10,
				"Float":  10.10,
				"String": "str",
			}

			Expect(data).To(Equal(d))
		})

		It("Should ignore empty fields on ToMap", func() {
			s := nestedStruct{
				testStruct: testStruct{
					Int:    10,
					Float:  10.10,
					String: "str",
				},
				Strs: make([]string, 0),
			}

			data := Reflect(s).MustStruct().ToMap(true, true)

			d := map[string]interface{}{
				"Int":    10,
				"Float":  10.10,
				"String": "str",
			}

			Expect(data).To(Equal(d))
		})

		It("Should load data from map", func() {
			d := map[string]interface{}{
				"Int":    10,
				"Float":  10.10,
				"String": "str",

				"Embedded": map[string]interface{}{
					"Int":    20,
					"Float":  20.20,
					"String": "str2",
				},

				"EmbeddedPtr": map[string]interface{}{
					"Int":    30,
					"Float":  30.30,
					"String": "str3",
				},
			}

			cs := nestedStruct{
				testStruct: testStruct{
					Int:    10,
					Float:  10.10,
					String: "str",
				},

				Embedded: testStruct{
					Int:    20,
					Float:  20.20,
					String: "str2",
				},

				EmbeddedPtr: &testStruct{
					Int:    30,
					Float:  30.30,
					String: "str3",
				},
			}

			s := &nestedStruct{}
			r := Reflect(s).MustStruct()
			Expect(r.FromMap(d)).ToNot(HaveOccurred())

			Expect(*s).To(Equal(cs))
		})
	})
})
