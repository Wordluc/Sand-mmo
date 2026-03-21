package common

import (
	"cmp"
	"slices"
	"sync"
)

type OrderList[t cmp.Ordered] struct {
	list []t
	*sync.Mutex
}

func NewOrderList[t cmp.Ordered]() OrderList[t] {
	return OrderList[t]{
		list:  make([]t, 0),
		Mutex: &sync.Mutex{},
	}
}

func (a *OrderList[t]) SortedInsert(newValue t) {
	a.Mutex.Lock()
	i, found := slices.BinarySearch(a.list, newValue)
	if !found {
		a.list = slices.Insert(a.list, i, newValue)
	}
	a.Mutex.Unlock()
}

func (a *OrderList[t]) Get() []t {
	return slices.Clone(a.list)
}

func (a *OrderList[t]) GetReversSort() []t {
	a.Lock()
	var res []t = make([]t, len(a.list))
	for i := range a.list {
		res[len(a.list)-i-1] = (a.list)[i]
	}
	a.Unlock()
	return res
}

func (a *OrderList[t]) Clean() {
	a.Lock()
	a.list = make([]t, 0)
	a.Unlock()
}
