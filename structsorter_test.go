package reflector_test

import (
	. "github.com/theduke/go-reflector"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Structsorter", func() {
	It("Should sort int slice ascending", func() {
		s := []testStruct{
			testStruct{Int: 87},
			testStruct{Int: 1000},
			testStruct{Int: 5},
			testStruct{Int: 800},
			testStruct{Int: 2},
		}

		cs := []testStruct{
			testStruct{Int: 2},
			testStruct{Int: 5},
			testStruct{Int: 87},
			testStruct{Int: 800},
			testStruct{Int: 1000},
		}

		r := Reflect(s).MustSlice()
		Expect(SortStructSlice(r, "Int", true)).ToNot(HaveOccurred())

		Expect(s).To(Equal(cs))
	})

	It("Should sort int slice descending", func() {
		s := []testStruct{
			testStruct{Int: 87},
			testStruct{Int: 1000},
			testStruct{Int: 5},
			testStruct{Int: 800},
			testStruct{Int: 2},
		}

		cs := []testStruct{
			testStruct{Int: 1000},
			testStruct{Int: 800},
			testStruct{Int: 87},
			testStruct{Int: 5},
			testStruct{Int: 2},
		}

		r := Reflect(s).MustSlice()
		Expect(SortStructSlice(r, "Int", false)).ToNot(HaveOccurred())

		Expect(s).To(Equal(cs))
	})
})
