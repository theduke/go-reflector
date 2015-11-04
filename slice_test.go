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

		filterFunc := func(item *Reflector) (bool, error) {
			return item.Interface().(int)%2 == 0, nil
		}

		newSlice, err := r.FilterBy(filterFunc)
		Expect(err).ToNot(HaveOccurred())
		Expect(newSlice.Interface().([]int)).To(Equal([]int{0, 2, 4, 6, 8, 10}))
	})

	It("Should sort slice with .SortBy()", func() {
		s := []int{3, 10, 8, 2, 5, 7, 1, 4, 0, 9, 6}
		r := R(s).MustSlice()

		sorter := func(a, b *Reflector) (bool, error) {
			return a.Interface().(int) < b.Interface().(int), nil
		}

		Expect(r.SortBy(sorter)).ToNot(HaveOccurred())
		Expect(s).To(Equal([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))
	})

	It("Should sort by field with .SortByFieldFunc() with maps", func() {
		items := []map[string]interface{}{
			map[string]interface{}{"int": 10},
			map[string]interface{}{"int": 2},
			map[string]interface{}{"int": 6},
			map[string]interface{}{"int": 8},
			map[string]interface{}{"int": 0},
			map[string]interface{}{"int": 4},
		}

		sortedItems := []map[string]interface{}{
			map[string]interface{}{"int": 0},
			map[string]interface{}{"int": 2},
			map[string]interface{}{"int": 4},
			map[string]interface{}{"int": 6},
			map[string]interface{}{"int": 8},
			map[string]interface{}{"int": 10},
		}

		r := R(items).MustSlice()

		sorter := func(a, b *Reflector) (bool, error) {
			return a.Interface().(int) < b.Interface().(int), nil
		}

		Expect(r.SortByFieldFunc("int", sorter)).ToNot(HaveOccurred())
		Expect(items).To(Equal(sortedItems))
	})

	It("Should sort by field with .SortByFieldFunc() with structs", func() {
		type S struct{ Int int }
		items := []S{
			S{10},
			S{2},
			S{6},
			S{8},
			S{0},
			S{4},
		}
		sortedItems := []S{
			S{0},
			S{2},
			S{4},
			S{6},
			S{8},
			S{10},
		}

		r := R(items).MustSlice()

		sorter := func(a, b *Reflector) (bool, error) {
			return a.Interface().(int) < b.Interface().(int), nil
		}

		Expect(r.SortByFieldFunc("Int", sorter)).ToNot(HaveOccurred())
		Expect(items).To(Equal(sortedItems))
	})

	It("Should sort by field with .SortByFieldFunc() with struct pointers", func() {
		type S struct{ Int int }
		items := []*S{
			&S{10},
			&S{2},
			&S{6},
			&S{8},
			&S{0},
			&S{4},
		}
		sortedItems := []*S{
			&S{0},
			&S{2},
			&S{4},
			&S{6},
			&S{8},
			&S{10},
		}

		r := R(items).MustSlice()

		sorter := func(a, b *Reflector) (bool, error) {
			return a.Interface().(int) < b.Interface().(int), nil
		}

		Expect(r.SortByFieldFunc("Int", sorter)).ToNot(HaveOccurred())
		Expect(items).To(BeEquivalentTo(sortedItems))
	})

	It("Should sort by field with .SortByField()", func() {
		type S struct{ Int int }
		items := []*S{
			&S{10},
			&S{2},
			&S{6},
			&S{8},
			&S{0},
			&S{4},
		}
		sortedItems := []*S{
			&S{0},
			&S{2},
			&S{4},
			&S{6},
			&S{8},
			&S{10},
		}

		r := R(items).MustSlice()

		Expect(r.SortByField("Int", true)).ToNot(HaveOccurred())
		Expect(items).To(BeEquivalentTo(sortedItems))
	})

	It("Should sort by field with .SortByField() descending", func() {
		type S struct{ Int int }
		items := []*S{
			&S{10},
			&S{2},
			&S{6},
			&S{8},
			&S{0},
			&S{4},
		}
		sortedItems := []*S{
			&S{10},
			&S{8},
			&S{6},
			&S{4},
			&S{2},
			&S{0},
		}

		r := R(items).MustSlice()

		Expect(r.SortByField("Int", false)).ToNot(HaveOccurred())
		Expect(items).To(BeEquivalentTo(sortedItems))
	})
})
