package reflector

import (
	"errors"
	"sort"
)

/**
- * Sorter for sorting structs by field.
- */

type structFieldSorter struct {
	slice    *SliceReflector
	fields   []*Reflector
	operator string

	err error
}

func (s structFieldSorter) Len() int {
	return s.slice.Len()
}

func (s structFieldSorter) Swap(i, j int) {
	s.slice.Swap(i, j)
}

func (s structFieldSorter) Less(i, j int) bool {
	iField := s.fields[i]
	jField := s.fields[j]

	flag, err := iField.CompareTo(jField, s.operator)
	if err != nil {
		s.err = err
		return true
	}

	return flag
}

func newStructFieldSorter(slice *SliceReflector, fieldName string, asc bool) (*structFieldSorter, error) {
	if slice.Len() < 1 {
		return nil, errors.New("empty_slice")
	}

	fields := make([]*Reflector, 0, slice.Cap())

	for _, item := range slice.Items() {
		r, err := item.Struct()
		if err != nil {
			return nil, err
		}
		field := r.Field(fieldName)
		if field == nil {
			return nil, errors.New(ERR_UNKNOWN_FIELD)
		}
		fields = append(fields, field)
	}

	operator := "<"
	if !asc {
		operator = ">"
	}

	// Check that field can be compared.
	if _, err := fields[0].CompareTo(fields[0], operator); err != nil {
		return nil, errors.New(ERR_UNCOMPARABLE_VALUES)
	}

	return &structFieldSorter{
		slice:    slice,
		fields:   fields,
		operator: operator,
	}, nil
}

func SortStructSlice(items *SliceReflector, field string, ascending bool) error {
	sorter, err := newStructFieldSorter(items, field, ascending)
	if err != nil {
		return err
	}

	sort.Sort(*sorter)
	if sorter.err != nil {
		return sorter.err
	}
	return nil
}
