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

	It("Should .Cap()", func() {
		s := make([]int, 0, 10)
		Expect(Reflect(s).MustSlice().Cap()).To(Equal(10))
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

	It("Should set index", func() {
		s := []int{0, 1, 2}
		r := Reflect(s).MustSlice()
		err := r.SetIndexValue(0, 55)
		Expect(err).ToNot(HaveOccurred())
		Expect(s[0]).To(Equal(55))
	})

	It("Should .Swap() indexes", func() {
		s := []int{0, 1, 2}
		r := Reflect(s).MustSlice()
		err := r.Swap(0, 2)
		Expect(err).ToNot(HaveOccurred())
		Expect(s).To(Equal([]int{2, 1, 0}))
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

	It("Should filter slice with .FilterBy()", func() {
		s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		r := R(s).MustSlice()
		newSlice := r.FilterBy(func(item *Reflector) bool {
			return item.Interface().(int)%2 == 0
		}).Interface().([]int)

		Expect(newSlice).To(Equal([]int{0, 2, 4, 6, 8, 10}))
	})
})
