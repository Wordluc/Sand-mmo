package common

import (
	"cmp"
	"slices"
)

type OrderList[t cmp.Ordered] []t

func (a *OrderList[t]) SortedInsert(newValue t) {
	i, found := slices.BinarySearch(*a, newValue)
	if found {
		return
	}
	*a = slices.Insert(*a, i, newValue)
}

func (a *OrderList[t]) GetReversSort() []t {
	var res []t = make([]t, len(*a))
	for i := range *a {
		res[len(*a)-i-1] = (*a)[i]
	}
	return res
}

func (a *OrderList[t]) Clean() {

	*a = []t{}
}
