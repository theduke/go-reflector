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

	It("Should create SliceReflector from empty slice ptr", func() {
		var s []int
		r, err := Reflect(&s).Slice()
		Expect(err).ToNot(HaveOccurred())
		Expect(r.Len()).To(Equal(0))
	})

	It("Should .Len()", func() {
		s := []int{0, 1, 2}
		r, _ := Reflect(s).Slice()
		Expect(r.Len()).To(Equal(3))
	})

	It("Should return item with .Index()", func() {
		s := []int{0, 1, 2}
		r, _ := Reflect(s).Slice()
		Expect(r.Index(1).Interface()).To(Equal(1))
	})

	It("Should return nil with .Index() for inexistant index", func() {
		s := []int{0, 1, 2}
		r, _ := Reflect(s).Slice()
		Expect(r.Len()).To(Equal(3))
		Expect(r.Index(5)).To(BeNil())
	})

	It("Should return items with .Items()", func() {
		r := Reflect([]int{1, 2, 3}).MustSlice()
		items := r.Items()

		Expect(items).To(HaveLen(3))
		Expect(items[0].Interface()).To(Equal(1))
		Expect(items[1].Interface()).To(Equal(2))
		Expect(items[2].Interface()).To(Equal(3))
	})

	It("Should .Append() and AppendValue()", func() {
		s := []int{}
		r, _ := Reflect(&s).Slice()
		Expect(r.Append(Reflect(5))).ToNot(HaveOccurred())
		Expect(r.Len()).To(Equal(1))
		Expect(len(s)).To(Equal(1))
		Expect(r.IndexValue(0)).To(Equal(5))
		Expect(s[0]).To(Equal(5))

		Expect(r.AppendValue(10, 15, 20)).ToNot(HaveOccurred())
		Expect(s).To(Equal([]int{5, 10, 15, 20}))
	})

	It("Should convert interface slice to int", func() {
		s := []interface{}{0, 1, 2, 3}
		Expect(Reflect(s).MustSlice().ConvertTo(0)).To(Equal([]int{0, 1, 2, 3}))
	})

	It("Should convert int slice to float", func() {
		s := []interface{}{0, 1, 2, 3}
		Expect(Reflect(s).MustSlice().ConvertTo(float64(0))).To(Equal([]float64{float64(0), float64(1), float64(2), float64(3)}))
	})
})
