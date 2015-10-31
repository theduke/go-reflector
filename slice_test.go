package reflector_test

import (
	. "github.com/theduke/go-reflector"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Slice", func() {

	It("Should create SliceReflector from slice", func() {
		s, err := Reflect([]int{1}).Slice()
		Expect(err).ToNot(HaveOccurred())
		Expect(s.Len()).To(Equal(1))
	})

	It("Should create SliceReflector from slice ptr", func() {
		s := []int{1}
		r, err := Reflect(&s).Slice()
		Expect(err).ToNot(HaveOccurred())
		Expect(r.Len()).To(Equal(1))
	})

	It("Should create SliceReflector from empty ptr", func() {
		var s *[]int
		r, err := Reflect(&s).Slice()
		Expect(err).ToNot(HaveOccurred())
		Expect(Reflect(s).IsNil()).To(BeFalse())
		Expect(r.Len()).To(Equal(0))
	})
})
